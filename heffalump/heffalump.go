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
// buffer/io operations. This is called during package init before config
// is loaded, so it must NOT read any config values. Config-dependent setup
// (chunk pool, rate limiter) is deferred to InitFromConfig().
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

	return h
}

// InitFromConfig completes Heffalump initialization using config values that
// are not available during package init (pool, rate limiter). Must be called
// after config.Init() and config.StartLogger(). Also re-fetches the logger
// since the package-level var captured the pre-init fallback.
func InitFromConfig() {
	// The package-level log was set during package init before the real logger
	// existed. Re-fetch now that StartLogger has been called.
	log = config.GetLogger()

	// Initialise chunk pool if configured.
	if config.Perf.Chunks.PoolSizeMB > 0 {
		log.Info().
			Int("pool_mb", config.Perf.Chunks.PoolSizeMB).
			Int("chunk_kb", config.Perf.Chunks.ChunkSizeKB).
			Int("refill_kbps", config.Perf.Chunks.RefillRateKbps).
			Msg("Pre-generating Markov chunk pool — this may take a moment on constrained hardware...")
		GlobalPool = NewChunkPool(
			config.Perf.Chunks.PoolSizeMB,
			config.Perf.Chunks.ChunkSizeKB,
			config.Perf.Chunks.RefillRateKbps,
			DefaultHeffalump.mm,
		)
		chunkCount := (config.Perf.Chunks.PoolSizeMB * 1024 * 1024) / (config.Perf.Chunks.ChunkSizeKB * 1024)
		log.Info().
			Int("chunks", chunkCount).
			Int("chunk_kb", config.Perf.Chunks.ChunkSizeKB).
			Msg("Chunk pool ready")
	}

	// Initialise global rate limiter if configured.
	if config.Perf.MaxTotalKbps > 0 {
		initGlobalRateLimiter(config.Perf.MaxTotalKbps)
		log.Info().
			Int("max_total_kbps", config.Perf.MaxTotalKbps).
			Msg("Global rate limiter active")
	}
}

// writeSliceSize is the maximum number of bytes written per write-cycle when
// rate limiting is active. Keeping this small (4 KB) ensures:
//
//  1. Smooth data flow — the client receives a steady trickle instead of
//     long silence followed by large bursts.
//  2. Prompt disconnect detection — a broken connection surfaces as a write
//     error within one slice-interval rather than after an entire chunk's
//     worth of sleeping.
//  3. Accurate rate limiting — sleep durations are short and proportional,
//     so the actual throughput closely tracks the configured rate.
//
// When rate limiting is disabled (BaselineRateKbps == 0 and MaxTotalKbps == 0),
// full-chunk writes are used for maximum throughput.
const writeSliceSize = 4096

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
// Data is written in small slices (writeSliceSize) when rate limiting is active,
// ensuring prompt disconnect detection and smooth throughput. Without rate limiting,
// full-buffer writes are used for maximum speed.
func (h *Heffalump) WriteHell(bw *bufio.Writer) (int64, error) {
	var n int64

	defer func() {
		if r := recover(); r != nil {
			log.Error().Interface("caller", r).Msg("panic recovered!")
		}
	}()

	// Write and flush the HTML prefix immediately so the client sees data
	// before the first rate-limit sleep. Without the flush, the prefix sits
	// in the bufio buffer for the entire first sleep interval.
	if _, err := bw.WriteString("<html>\n<body>\n"); err != nil {
		return n, err
	}
	if err := bw.Flush(); err != nil {
		return n, err
	}

	rateLimited := config.Perf.BaselineRateKbps > 0 || globalRateBytes.Load() > 0

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
	if config.Perf.BaselineRateKbps > 0 {
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
			// When rate limiting is active, drain the buffer in small slices
			// so that (a) each sleep is short and proportional, (b) writes
			// happen frequently enough to detect client disconnects promptly,
			// and (c) the client receives a smooth trickle of data.
			//
			// When rate limiting is disabled, write the entire buffer at once
			// for maximum throughput — there is nothing to sleep for, and
			// disconnect detection is immediate on the next write anyway.
			if rateLimited {
				if err := h.writeSliced(bw, buf[:nr], &n, writeStart); err != nil {
					return n, err
				}
			} else {
				if _, ew := bw.Write(buf[:nr]); ew != nil {
					return n, ew
				}
				n += int64(nr)
			}
		}
	}
}

// writeSliced drains data in writeSliceSize increments, applying per-connection
// and global rate limiting between each slice. total is updated in place so the
// caller's byte counter stays current across slices.
func (h *Heffalump) writeSliced(bw *bufio.Writer, data []byte, total *int64, writeStart time.Time) error {
	written := 0
	for written < len(data) {
		end := written + writeSliceSize
		if end > len(data) {
			end = len(data)
		}
		sliceLen := int64(end - written)

		// Per-connection rate limiting: sleep if we are ahead of the target rate.
		if config.Perf.BaselineRateKbps > 0 {
			bytesPerSec := int64(config.Perf.BaselineRateKbps) * 1024
			expected := time.Duration(float64(time.Second) * float64(*total+int64(written)+sliceLen) / float64(bytesPerSec))
			if sleep := expected - time.Since(writeStart); sleep > 0 {
				time.Sleep(sleep)
			}
		}

		// Global rate limiting: wait for token budget before writing.
		if globalRateBytes.Load() > 0 {
			globalRateWait(sliceLen)
		}

		if _, ew := bw.Write(data[written:end]); ew != nil {
			*total += int64(written)
			return ew
		}
		written = end
	}
	*total += int64(written)
	return nil
}
