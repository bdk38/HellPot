package heffalump

import (
	"sync"
	"sync/atomic"
	"time"
)

// ChunkPool holds a fixed set of pre-generated Markov text chunks.
// Connections read chunks sequentially in round-robin order, copying each
// chunk into the write buffer (memcpy) instead of generating Markov text
// on the fly. This moves almost all CPU cost from the per-connection hot
// path to a single background goroutine that regenerates chunks lazily at
// a controlled rate.
//
// Each chunk is protected by its own RWMutex. Many connections can read
// the same chunk simultaneously; the refill goroutine takes an exclusive
// write lock only while swapping a single chunk, leaving all others
// readable.
type ChunkPool struct {
	chunks    [][]byte
	mu        []sync.RWMutex
	idx       atomic.Uint64
	count     uint64 // len(chunks) as uint64; keeps CopyChunk arithmetic in uint64 space
	mm        MarkovMap
	ChunkSize int // exported so WriteHell can size its copy buffer
}

// GlobalPool is the package-level chunk pool. It is nil when
// pool_size_mb = 0 (the default), in which case WriteHell falls back to
// on-the-fly Markov generation. Set by NewHeffalump when pool is enabled.
var GlobalPool *ChunkPool

// NewChunkPool allocates and fills the pool. poolSizeMB and chunkSizeKB
// determine count and size; refillRateKbps governs the background goroutine.
// This call blocks until all chunks are generated — log a message before
// calling on large pools so the operator sees what is happening.
func NewChunkPool(poolSizeMB, chunkSizeKB, refillRateKbps int, mm MarkovMap) *ChunkPool {
	chunkSize := chunkSizeKB * 1024
	count := uint64((poolSizeMB * 1024 * 1024) / chunkSize) // #nosec G115 -- poolSizeMB and chunkSize are config-validated positive values
	if count < 1 {
		count = 1
	}

	p := &ChunkPool{
		chunks:    make([][]byte, count),
		mu:        make([]sync.RWMutex, count),
		count:     count,
		mm:        mm,
		ChunkSize: chunkSize,
	}

	for i := range p.chunks {
		p.chunks[i] = p.generate()
	}

	go p.refillLoop(refillRateKbps)
	return p
}

// generate produces one full chunk of Markov text.
//
// The loop terminates when the buffer is full OR when Read returns 0 bytes
// (the remaining space is too small for any word + separator). A chunk that
// is a few bytes short is functionally identical to a perfectly full one —
// bots cannot tell the difference — so we do not spin trying to find a
// word that fits the last 1-2 bytes.
func (p *ChunkPool) generate() []byte {
	buf := make([]byte, p.ChunkSize)
	mr := NewMarkovReader(p.mm)
	total := 0
	for total < p.ChunkSize {
		n, err := mr.Read(buf[total:])
		total += n
		if n == 0 || err != nil || total >= p.ChunkSize {
			break
		}
	}
	return buf[:total]
}

// CopyChunk copies the next chunk in round-robin order into dst and returns
// the number of bytes copied. The modulo uses the pre-stored uint64 count to
// keep the arithmetic entirely in uint64 space — avoiding the integer overflow
// conversion that would otherwise be flagged by static analysis (G115/CWE-190).
// The copy releases the read lock before returning, so the refill goroutine is
// never blocked by a slow writer.
func (p *ChunkPool) CopyChunk(dst []byte) int {
	i := p.idx.Add(1) % p.count
	p.mu[i].RLock()
	n := copy(dst, p.chunks[i])
	p.mu[i].RUnlock()
	return n
}

// refillLoop regenerates one chunk at a time at a rate derived from
// refillRateKbps, cycling through the pool indefinitely. Running slowly
// keeps the background CPU cost negligible on constrained hardware while
// ensuring the pool never serves permanently stale content.
func (p *ChunkPool) refillLoop(refillRateKbps int) {
	if refillRateKbps <= 0 {
		refillRateKbps = 128
	}
	// How long to pause between chunk regenerations to maintain the target rate.
	refillBPS := int64(refillRateKbps) * 1024
	chunkDur := time.Duration(float64(time.Second) * float64(p.ChunkSize) / float64(refillBPS))

	i := 0
	for {
		time.Sleep(chunkDur)
		fresh := p.generate()
		p.mu[i].Lock()
		p.chunks[i] = fresh
		p.mu[i].Unlock()
		i = (i + 1) % len(p.chunks)
	}
}
