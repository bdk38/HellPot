package heffalump

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"math/rand/v2"
	"strings"
	"unicode"
	"unicode/utf8"
)

var DefaultMarkovMap MarkovMap

// unpackGzipBase64 decodes a base64-encoded gzipped string using Go's standard library.
func unpackGzipBase64(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	reader, err := gzip.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return "", err
	}
	defer reader.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func init() {
	// DefaultMarkovMap is a Markov chain based on src.
	src, err := unpackGzipBase64(srcGz)
	if err != nil {
		panic(err)
	}
	if len(src) < 1 {
		panic("failed to unpack source")
	}
	DefaultMarkovMap = MakeMarkovMap(strings.NewReader(src))
	DefaultHeffalump = NewHeffalump(DefaultMarkovMap, 100*1<<10)
}

// ScanHTML is a basic split function for a Scanner that returns each
// space-separated word of text or HTML tag, with surrounding spaces deleted.
// It will never return an empty string. The definition of space is set by
// unicode.IsSpace.
func ScanHTML(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip leading spaces.
	var r rune
	var start = 0
	for width := 0; start < len(data); start += width {
		r, width = utf8.DecodeRune(data[start:])
		if !unicode.IsSpace(r) {
			break
		}
	}
	switch {
	case r == '<':
		// Scan until closing bracket
		for i := start; i < len(data); i++ {
			if data[i] == '>' {
				return i + 1, data[start : i+1], nil
			}
		}
	default:
		// Scan until space, marking end of word.
		for width, i := 0, start; i < len(data); i += width {
			var r rune
			r, width = utf8.DecodeRune(data[i:])
			if unicode.IsSpace(r) {
				return i + width, data[start:i], nil
			}
			if r == '<' {
				return i, data[start:i], nil
			}
		}
	}
	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}
	// Request more data.
	return start, nil, nil
}

type tokenPair [2]string

// MarkovMap is a map that acts as a Markov chain generator.
type MarkovMap map[tokenPair][]string

// MakeMarkovMap makes an empty MakeMarkov and fills it with r.
func MakeMarkovMap(r io.Reader) MarkovMap {
	m := MarkovMap{}
	m.Fill(r)
	return m
}

// Fill adds all the tokens in r to a MarkovMap
func (mm MarkovMap) Fill(r io.Reader) {
	var w1, w2, w3 string

	s := bufio.NewScanner(r)
	s.Split(ScanHTML)
	for s.Scan() {
		w3 = s.Text()
		mm.Add(w1, w2, w3)
		w1, w2 = w2, w3
	}

	mm.Add(w1, w2, w3)
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
	r := rand.IntN(len(suffix))
	return suffix[r]
}

// Read fills p with data from calling Get on the MarkovMap.
func (mm MarkovMap) Read(p []byte) (n int, err error) {
	var w1, w2, w3 string
	for {
		w3 = mm.Get(w1, w2)
		if n+len(w3)+1 >= len(p) {
			break
		}
		n += copy(p[n:], w3)
		n += copy(p[n:], "\n")
		w1, w2 = w2, w3
	}
	return
}
