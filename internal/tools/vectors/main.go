// Command vectors builds question/data/vectors.bin from pre-trained GloVe
// word vectors. It's run by hand when the vectors change, not as part of the
// build; its output is committed.
//
// It reads the 100-dimensional GloVe 6B vectors (Wikipedia 2014 + Gigaword
// 5; public domain under the PDDL) in word2vec text format, as published by
// gensim-data:
//
//	curl -LO https://github.com/RaRe-Technologies/gensim-data/releases/download/glove-wiki-gigaword-100/glove-wiki-gigaword-100.gz
//	go run ./internal/tools/vectors glove-wiki-gigaword-100.gz question/data/vectors.bin
//
// It keeps the most frequent purely alphabetic words, scales each vector to
// unit length and quantizes it to int8, so the dot product of two stored
// vectors, divided by 127², is their cosine similarity.
//
// The output format, little-endian: "GMV1"; dimensions and word count as
// uint32; the words, each a uint8 length then its bytes; then the vectors,
// count × dimensions int8.
package main

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

func main() {
	words := flag.Int("words", 50000, "number of words to keep")
	flag.Parse()
	if flag.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "usage: vectors [-words N] INPUT.gz OUTPUT")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), flag.Arg(1), *words); err != nil {
		fmt.Fprintln(os.Stderr, "vectors:", err)
		os.Exit(1)
	}
}

func alphabetic(w string) bool {
	if len(w) == 0 || len(w) > 255 {
		return false
	}
	for _, r := range w {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	return true
}

func run(in, out string, limit int) error {
	f, err := os.Open(in)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	sc := bufio.NewScanner(gz)
	sc.Buffer(make([]byte, 1<<20), 1<<20)

	if !sc.Scan() {
		return errors.New("empty input")
	}
	header := strings.Fields(sc.Text())
	if len(header) != 2 {
		return fmt.Errorf("bad header %q", sc.Text())
	}
	dims, err := strconv.Atoi(header[1])
	if err != nil {
		return err
	}

	var vocab []string
	var vecs []int8
	for sc.Scan() && len(vocab) < limit {
		fields := strings.Fields(sc.Text())
		if len(fields) != dims+1 || !alphabetic(fields[0]) {
			continue
		}
		v := make([]float64, dims)
		norm := 0.0
		for i, s := range fields[1:] {
			if v[i], err = strconv.ParseFloat(s, 64); err != nil {
				return err
			}
			norm += v[i] * v[i]
		}
		norm = math.Sqrt(norm)
		for _, x := range v {
			vecs = append(vecs, int8(math.Round(x/norm*127)))
		}
		vocab = append(vocab, fields[0])
	}
	if err := sc.Err(); err != nil {
		return err
	}

	o, err := os.Create(out)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(o)
	if err := write(w, dims, vocab, vecs); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}
	fmt.Printf("wrote %d words × %d dimensions to %s\n", len(vocab), dims, out)
	return o.Close()
}

func write(w io.Writer, dims int, vocab []string, vecs []int8) error {
	if _, err := io.WriteString(w, "GMV1"); err != nil {
		return err
	}
	for _, n := range []uint32{uint32(dims), uint32(len(vocab))} {
		if err := binary.Write(w, binary.LittleEndian, n); err != nil {
			return err
		}
	}
	for _, word := range vocab {
		if _, err := w.Write(append([]byte{byte(len(word))}, word...)); err != nil {
			return err
		}
	}
	return binary.Write(w, binary.LittleEndian, vecs)
}
