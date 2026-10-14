package export

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"io"

	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/compress/zstd"

	"github.com/storskegg/go-geomancer/chart"
)

// csvWriter writes a header row and then one row per record.
type csvWriter struct {
	w      *csv.Writer
	header bool
}

func newCSVWriter(w io.Writer) *csvWriter { return &csvWriter{w: csv.NewWriter(w)} }

func (c *csvWriter) Write(rec Record) error {
	if !c.header {
		if err := c.w.Write(Columns()); err != nil {
			return err
		}
		c.header = true
	}
	row := NewRow(rec)
	return c.w.Write(row.values())
}

func (c *csvWriter) Close() error {
	if !c.header {
		if err := c.w.Write(Columns()); err != nil {
			return err
		}
	}
	c.w.Flush()
	return c.w.Error()
}

// jsonlWriter writes one reading per line, as geomancer cast --json does,
// with the chart ID added.
type jsonlWriter struct {
	buf *bufio.Writer
	enc *json.Encoder
}

// jsonlRecord is the JSON Lines form of a record.
type jsonlRecord struct {
	Chart int `json:"chart"`
	chart.Reading
}

func newJSONLWriter(w io.Writer) *jsonlWriter {
	buf := bufio.NewWriter(w)
	return &jsonlWriter{buf: buf, enc: json.NewEncoder(buf)}
}

func (j *jsonlWriter) Write(rec Record) error {
	return j.enc.Encode(jsonlRecord{Chart: rec.Chart, Reading: rec.Reading})
}

func (j *jsonlWriter) Close() error { return j.buf.Flush() }

// parquetBatch is how many rows are buffered between writes.
const parquetBatch = 4096

// parquetWriter writes zstd-compressed Parquet with dictionary-encoded
// figure and sign columns.
type parquetWriter struct {
	w     *parquet.GenericWriter[Row]
	batch []Row
}

func newParquetWriter(w io.Writer) *parquetWriter {
	return &parquetWriter{
		w:     parquet.NewGenericWriter[Row](w, parquet.Compression(&zstd.Codec{})),
		batch: make([]Row, 0, parquetBatch),
	}
}

func (p *parquetWriter) Write(rec Record) error {
	p.batch = append(p.batch, NewRow(rec))
	if len(p.batch) == cap(p.batch) {
		return p.flush()
	}
	return nil
}

func (p *parquetWriter) flush() error {
	_, err := p.w.Write(p.batch)
	p.batch = p.batch[:0]
	return err
}

func (p *parquetWriter) Close() error {
	if err := p.flush(); err != nil {
		return err
	}
	return p.w.Close()
}
