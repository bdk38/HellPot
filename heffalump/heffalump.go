/*
Package heffalump attempts to encapsulate the original work by carlmjohnson on heffalump
https://github.com/carlmjohnson/heffalump
*/
package heffalump

import (
	"bufio"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bdk38/HellPot/internal/config"
)

var log = config.GetLogger()

// DefaultHeffalump represents a Heffalump type
var DefaultHeffalump *Heffalump

// Heffalump represents our buffer pool and markov map from Heffalump
type Heffalump struct {
	pool       *sync.Pool // byte buffer pool
	readerPool *sync.Pool // MarkovReader pool
	buffsize   int
	mm         MarkovMap
}

// Global rate limiter state. Initialised by NewHeffalump when MaxTotalKbps > 0.
// Uses a token-bucket approach: a background goroutine adds tokens on a 100ms
// ticker; WriteHell consumes tokens before each write and sleeps if the bucket
// is empty. atomic.Int64 keeps the hot path lock-free.
var (
	globalTokens    atomic.Int64
	globalRateBytes atomic.Int64 // bytes per second; 0 = disabled
)

// initGlobalRateLimiter starts the token bucket refill goroutine.
func initGlobalRateLimiter(kbps int) {
	bps := int64(kbps) * 1024
	globalRateBytes.Store(bps)
	globalTokens.Store(bps) // start with a full bucket
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			bps := globalRateBytes.Load()
			if bps <= 0 {
				return
			}
			add := bps / 10 // replenish 10% of capacity every 100ms
			for {
				cur := globalTokens.Load()
				want := cur + add
				if want > bps {
					want = bps
				}
				if globalTokens.CompareAndSwap(cur, want) {
					break
				}
			}
		}
	}()
}

// globalRateWait blocks until the global token bucket has n bytes available,
// consuming them atomically. No-op when the global rate limiter is disabled.
func globalRateWait(n int64) {
	bps := globalRateBytes.Load()
	if bps <= 0 {
		return
	}
	for {
		cur := globalTokens.Load()
		if cur >= n {
			if globalTokens.CompareAndSwap(cur, cur-n) {
				return
			}
			// CAS lost to another goroutine — retry immediately
			continue
		}
		// Insufficient tokens — sleep proportionally to the deficit
		deficit := n - cur
		sleep := time.Duration(float64(time.Second) * float64(deficit) / float64(bps))
		if sleep < time.Millisecond {
			sleep = time.Millisecond
		}
		time.Sleep(sleep)
	}
}

// NewHeffalump instantiates a new Heffalump for markov generation and
// buffer/io operations. It also initialises the chunk pool and global rate
// limiter based on the current config, so it must be called after
// config.Init() and config.StartLogger().
func NewHeffalump(mm MarkovMap, buffsize int) *Heffalump {
	h := &Heffalump{
		buffsize: buffsize,
		mm:       mm,
	}
	h.pool = &sync.Pool{New: func() any {
		return make([]byte, buffsize)
	}}
	h.readerPool = &sync.Pool{New: func() any {
		return NewMarkovReader(mm)
	}}

	// Initialise chunk pool if configured.
	if config.ChunkPoolSizeMB > 0 {
		log.Info().
			Int("pool_mb", config.ChunkPoolSizeMB).
			Int("chunk_kb", config.ChunkSizeKB).
			Int("refill_kbps", config.ChunkRefillRateKbps).
			Msg("Pre-generating Markov chunk pool — this may take a moment on constrained hardware...")
		GlobalPool = NewChunkPool(
			config.ChunkPoolSizeMB,
			config.ChunkSizeKB,
			config.ChunkRefillRateKbps,
			mm,
		)
		chunkCount := (config.ChunkPoolSizeMB * 1024 * 1024) / (config.ChunkSizeKB * 1024)
		log.Info().
			Int("chunks", chunkCount).
			Int("chunk_kb", config.ChunkSizeKB).
			Msg("Chunk pool ready")
	}

	// Initialise global rate limiter if configured.
	if config.MaxTotalKbps > 0 {
		initGlobalRateLimiter(config.MaxTotalKbps)
		log.Info().
			Int("max_total_kbps", config.MaxTotalKbps).
			Msg("Global rate limiter active")
	}

	return h
}

// WriteHell writes a continuous stream of Markov-generated text to bw.
//
// When a chunk pool is configured (GlobalPool != nil), it serves pre-generated
// chunks via memcpy instead of running the Markov chain inline — dramatically
// reducing per-connection CPU cost on constrained hardware.
//
// When BaselineRateKbps > 0, per-connection throughput is throttled by sleeping
// between writes to maintain the target rate. When MaxTotalKbps > 0, a global
// token bucket further limits aggregate outbound bandwidth across all connections.
//
// The <html><body> prefix has been removed; the HTTP skeleton is written by the
// router's trapSkeleton() before this function is called.
func (h *Heffalump) WriteHell(bw *bufio.Writer) (int64, error) {
	var n int64

	defer func() {
		if r := recover(); r != nil {
			log.Error().Interface("caller", r).Msg("panic recovered!")
		}
	}()

	if _, err := bw.WriteString("<html>\n<body>\n"); err != nil {
		return n, err
	}

	// When pool is enabled, allocate a copy buffer sized to one chunk and skip
	// the MarkovReader entirely. When pool is disabled, use the sync.Pool buffers
	// and a MarkovReader as before.
	var buf []byte
	var mr *MarkovReader

	if GlobalPool != nil {
		buf = make([]byte, GlobalPool.ChunkSize)
	} else {
		buf = h.pool.Get().([]byte)
		mr = h.readerPool.Get().(*MarkovReader)
		defer func() {
			h.pool.Put(buf)
			mr.reset()
			h.readerPool.Put(mr)
		}()
	}

	// Per-connection rate limiter: track when writing started so we can sleep
	// to maintain BaselineRateKbps KB/s. Uses elapsed-time comparison rather
	// than a per-connection token bucket to keep allocations minimal.
	var writeStart time.Time
	if config.BaselineRateKbps > 0 {
		writeStart = time.Now()
	}

	for {
		var nr int

		if GlobalPool != nil {
			nr = GlobalPool.CopyChunk(buf)
		} else {
			var er error
			nr, er = mr.Read(buf)
			if er != nil {
				return n, er
			}
		}

		if nr > 0 {
			// Per-connection rate limiting: sleep if we are ahead of the target rate.
			if config.BaselineRateKbps > 0 {
				bytesPerSec := int64(config.BaselineRateKbps) * 1024
				expected := time.Duration(float64(time.Second) * float64(n+int64(nr)) / float64(bytesPerSec))
				if sleep := expected - time.Since(writeStart); sleep > 0 {
					time.Sleep(sleep)
				}
			}

			// Global rate limiting: wait for token budget before writing.
			if globalRateBytes.Load() > 0 {
				globalRateWait(int64(nr))
			}

			if _, ew := bw.Write(buf[:nr]); ew != nil {
				return n, ew
			}
			n += int64(nr)
		}
	}
}
