package question

import (
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
)

// vectorFile holds GloVe word vectors (Wikipedia 2014 + Gigaword 5, public
// domain under the PDDL), trimmed to the most common words, scaled to unit
// length and quantized to int8. internal/tools/vectors builds it and
// documents the format.
//
//go:embed data/vectors.bin
var vectorFile []byte

// vectors maps words to their quantized unit vectors.
type vectors struct {
	dims  int
	index map[string]int
	data  []int8
}

var loadVectors = sync.OnceValues(func() (*vectors, error) { return parseVectors(vectorFile) })

func parseVectors(b []byte) (*vectors, error) {
	bad := errors.New("question: malformed vector file")
	if len(b) < 12 || string(b[:4]) != "GMV1" {
		return nil, bad
	}
	dims := int(binary.LittleEndian.Uint32(b[4:]))
	count := int(binary.LittleEndian.Uint32(b[8:]))
	b = b[12:]
	v := &vectors{dims: dims, index: make(map[string]int, count)}
	for i := range count {
		if len(b) == 0 || len(b) < 1+int(b[0]) {
			return nil, bad
		}
		n := int(b[0])
		v.index[string(b[1:1+n])] = i
		b = b[1+n:]
	}
	if len(b) != count*dims {
		return nil, fmt.Errorf("%w: %d bytes of vectors, want %d", bad, len(b), count*dims)
	}
	v.data = make([]int8, len(b))
	for i, x := range b {
		v.data[i] = int8(x)
	}
	return v, nil
}

// vector returns word's vector, or nil if it isn't known.
func (v *vectors) vector(word string) []int8 {
	i, ok := v.index[word]
	if !ok {
		return nil
	}
	return v.data[i*v.dims : (i+1)*v.dims]
}

// rank returns word's frequency rank, 0 for the most common, or -1 if it
// isn't known.
func (v *vectors) rank(word string) int {
	if i, ok := v.index[word]; ok {
		return i
	}
	return -1
}

// similarity returns the cosine similarity of two stored vectors.
func similarity(a, b []int8) float64 {
	dot := 0
	for i := range a {
		dot += int(a[i]) * int(b[i])
	}
	return float64(dot) / (127 * 127)
}
