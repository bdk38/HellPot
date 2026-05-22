package heffalump

import (
	"bufio"
	"io"
	"math/rand/v2"
	"strings"
)

var DefaultMarkovMap MarkovMap

func init() {
	// DefaultMarkovMap is a Markov chain based on src.
	DefaultMarkovMap = MakeMarkovMap(strings.NewReader(src))
	DefaultHeffalump = NewHeffalump(DefaultMarkovMap, 256*1<<10)
}

type tokenPair [2]string

// MarkovMap is a map that acts as a Markov chain generator.
type MarkovMap map[tokenPair][]string

// MakeMarkovMap makes an empty MarkovMap and fills it with r.
func MakeMarkovMap(r io.Reader) MarkovMap {
	m := MarkovMap{}
	m.Fill(r)
	return m
}

// Fill adds all the tokens in r to a MarkovMap.
//
// bufio.ScanWords is used in place of the previous ScanHTML split function.
// The corpus is now plain text so the HTML-aware scanner is unnecessary, and
// ScanWords avoids the per-byte utf8.DecodeRune overhead that ScanHTML paid
// even for ASCII-heavy input.
//
// All token strings are interned into a single map before being stored as
// bigram keys or successor values. This means every unique word is backed by
// exactly one string allocation. Subsequent map lookups hash the same pointer
// and length rather than re-hashing the underlying bytes on each Get call,
// which meaningfully reduces CPU overhead in the hot streaming path.
func (mm MarkovMap) Fill(r io.Reader) {
	intern := make(map[string]string)
	interned := func(s string) string {
		if v, ok := intern[s]; ok {
			return v
		}
		intern[s] = s
		return s
	}

	var w1, w2, w3 string

	s := bufio.NewScanner(r)
	s.Split(bufio.ScanWords)
	for s.Scan() {
		w3 = interned(s.Text())
		mm.Add(w1, w2, w3)
		w1, w2 = w2, w3
	}
	// Note: no trailing Add call here — the final trigram is already recorded
	// on the last iteration of the loop above. The previous extra call caused
	// the last token to appear twice as a successor, skewing its probability.
}

// Add adds a three token sequence to the map.
func (mm MarkovMap) Add(w1, w2, w3 string) {
	p := tokenPair{w1, w2}
	mm[p] = append(mm[p], w3)
}

// Get pseudo-randomly chooses a possible suffix to w1 and w2.
func (mm MarkovMap) Get(w1, w2 string) string {
	p := tokenPair{w1, w2}
	suffix, ok := mm[p]
	if !ok {
		return ""
	}
	// We don't care about cryptographically sound entropy here, ignore gosec G404.
	/* #nosec */
	return suffix[rand.IntN(len(suffix))]
}

// MarkovReader is a stateful io.Reader over a MarkovMap. Unlike calling
// Get directly, it carries w1/w2 across successive Read calls so that the
// generated token stream is one continuous chain walk per connection rather
// than a series of independent restarts from the empty-string bigram.
//
// MarkovReader values are intended to be pooled via sync.Pool in Heffalump;
// call reset() before returning one to the pool.
type MarkovReader struct {
	mm     MarkovMap
	w1, w2 string
}

// NewMarkovReader creates a MarkovReader for the given map, starting at the
// natural chain entry point (the empty-string bigram).
func NewMarkovReader(mm MarkovMap) *MarkovReader {
	return &MarkovReader{mm: mm}
}

// reset clears the chain state so the reader can be safely reused from the pool.
func (mr *MarkovReader) reset() {
	mr.w1, mr.w2 = "", ""
}

// Read fills p by walking the Markov chain. State is preserved across calls
// so each call continues the walk exactly where the previous one left off.
//
// On a dead-end (no successors for the current bigram), the walk resets to
// the empty-string entry point and continues — Read never returns an error
// or io.EOF, so the caller drives termination by closing the connection or
// stopping the copy loop.
//
// Tokens are separated by a single space written as a direct byte assignment,
// avoiding a second copy call per token.
func (mr *MarkovReader) Read(p []byte) (n int, err error) {
	for {
		w3 := mr.mm.Get(mr.w1, mr.w2)
		if w3 == "" {
			// Dead-end in the chain: reset to the entry point and keep going.
			mr.w1, mr.w2 = "", ""
			continue
		}
		// Ensure there is room for the token plus the trailing space separator.
		if n+len(w3)+1 > len(p) {
			break
		}
		n += copy(p[n:], w3)
		p[n] = ' '
		n++
		mr.w1, mr.w2 = mr.w2, w3
	}
	return
}
