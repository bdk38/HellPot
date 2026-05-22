/*
Package heffalump attempts to encapsulate the original work by carlmjohnson on heffalump
https://github.com/carlmjohnson/heffalump
*/
package heffalump

import (
	"bufio"
	"sync"

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

// NewHeffalump instantiates a new Heffalump for markov generation and buffer/io operations.
// Two pools are maintained: one for the raw byte buffers used by io.CopyBuffer, and one for
// MarkovReader values so that each connection gets a stateful reader whose chain walk persists
// across successive Read calls without allocating a new reader per call.
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

// WriteHell writes a continuous stream of Markov-generated text to bw.
//
// It acquires a MarkovReader from the reader pool so that the chain walk is
// stateful across the entire connection lifetime — every Read call continues
// where the last one left off rather than restarting from the empty bigram.
//
// Both the byte buffer and the MarkovReader are returned to their respective
// pools when WriteHell returns, with the reader state reset so it is ready
// for the next connection.
//
// Note: we use an explicit read/write loop rather than io.CopyBuffer.
// *bufio.Writer implements io.ReaderFrom, so io.CopyBuffer would call
// bw.ReadFrom(mr) which chains down through fasthttp's internal writer and
// returns after a single read cycle — producing a finite ~4KB response
// instead of an infinite stream. The explicit loop calls bw.Write directly,
// bypassing the ReaderFrom dispatch entirely.
func (h *Heffalump) WriteHell(bw *bufio.Writer) (int64, error) {
	var n int64

	defer func() {
		if r := recover(); r != nil {
			log.Error().Interface("caller", r).Msg("panic recovered!")
		}
	}()

	buf := h.pool.Get().([]byte)
	mr := h.readerPool.Get().(*MarkovReader)

	defer func() {
		h.pool.Put(buf)
		mr.reset()
		h.readerPool.Put(mr)
	}()

	if _, err := bw.WriteString("<html>\n<body>\n"); err != nil {
		return n, err
	}

	for {
		nr, er := mr.Read(buf)
		if nr > 0 {
			_, ew := bw.Write(buf[:nr])
			if ew != nil {
				return n, ew
			}
			n += int64(nr)
		}
		if er != nil {
			return n, er
		}
	}
}
