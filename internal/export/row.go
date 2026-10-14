package export

import (
	"strconv"
	"strings"

	"github.com/storskegg/go-geomancer/position"
)

// Row is one reading flattened into columns, for the CSV and Parquet
// formats. Field order is column order.
type Row struct {
	System string `parquet:"system,dict"`
	Chart  int32  `parquet:"chart"`
	// The Shield.
	M1 string `parquet:"m1,dict"`
	M2 string `parquet:"m2,dict"`
	M3 string `parquet:"m3,dict"`
	M4 string `parquet:"m4,dict"`
	D1 string `parquet:"d1,dict"`
	D2 string `parquet:"d2,dict"`
	D3 string `parquet:"d3,dict"`
	D4 string `parquet:"d4,dict"`
	N1 string `parquet:"n1,dict"`
	N2 string `parquet:"n2,dict"`
	N3 string `parquet:"n3,dict"`
	N4 string `parquet:"n4,dict"`
	WR string `parquet:"wr,dict"`
	WL string `parquet:"wl,dict"`
	J  string `parquet:"j,dict"`
	JM string `parquet:"jm,dict"`
	// The figure in each house.
	H1Figure  string `parquet:"h1_figure,dict"`
	H2Figure  string `parquet:"h2_figure,dict"`
	H3Figure  string `parquet:"h3_figure,dict"`
	H4Figure  string `parquet:"h4_figure,dict"`
	H5Figure  string `parquet:"h5_figure,dict"`
	H6Figure  string `parquet:"h6_figure,dict"`
	H7Figure  string `parquet:"h7_figure,dict"`
	H8Figure  string `parquet:"h8_figure,dict"`
	H9Figure  string `parquet:"h9_figure,dict"`
	H10Figure string `parquet:"h10_figure,dict"`
	H11Figure string `parquet:"h11_figure,dict"`
	H12Figure string `parquet:"h12_figure,dict"`
	// The sign of each house; nil when the system gives House 1 no sign.
	H1Sign  *string `parquet:"h1_sign,optional,dict"`
	H2Sign  *string `parquet:"h2_sign,optional,dict"`
	H3Sign  *string `parquet:"h3_sign,optional,dict"`
	H4Sign  *string `parquet:"h4_sign,optional,dict"`
	H5Sign  *string `parquet:"h5_sign,optional,dict"`
	H6Sign  *string `parquet:"h6_sign,optional,dict"`
	H7Sign  *string `parquet:"h7_sign,optional,dict"`
	H8Sign  *string `parquet:"h8_sign,optional,dict"`
	H9Sign  *string `parquet:"h9_sign,optional,dict"`
	H10Sign *string `parquet:"h10_sign,optional,dict"`
	H11Sign *string `parquet:"h11_sign,optional,dict"`
	H12Sign *string `parquet:"h12_sign,optional,dict"`
}

// shield returns the Shield fields, indexed by position.Position.
func (r *Row) shield() [16]*string {
	return [16]*string{
		&r.M1, &r.M2, &r.M3, &r.M4, &r.D1, &r.D2, &r.D3, &r.D4,
		&r.N1, &r.N2, &r.N3, &r.N4, &r.WR, &r.WL, &r.J, &r.JM,
	}
}

// houseFigures returns the house figure fields; index i is House i+1.
func (r *Row) houseFigures() [12]*string {
	return [12]*string{
		&r.H1Figure, &r.H2Figure, &r.H3Figure, &r.H4Figure, &r.H5Figure, &r.H6Figure,
		&r.H7Figure, &r.H8Figure, &r.H9Figure, &r.H10Figure, &r.H11Figure, &r.H12Figure,
	}
}

// houseSigns returns the house sign fields; index i is House i+1.
func (r *Row) houseSigns() [12]**string {
	return [12]**string{
		&r.H1Sign, &r.H2Sign, &r.H3Sign, &r.H4Sign, &r.H5Sign, &r.H6Sign,
		&r.H7Sign, &r.H8Sign, &r.H9Sign, &r.H10Sign, &r.H11Sign, &r.H12Sign,
	}
}

// Columns returns the column names, in order.
func Columns() []string {
	cols := []string{"system", "chart"}
	for _, p := range position.All() {
		cols = append(cols, strings.ToLower(p.String()))
	}
	for h := 1; h <= 12; h++ {
		cols = append(cols, "h"+strconv.Itoa(h)+"_figure")
	}
	for h := 1; h <= 12; h++ {
		cols = append(cols, "h"+strconv.Itoa(h)+"_sign")
	}
	return cols
}

// values returns the row as strings, in column order. A nil sign is "".
func (r *Row) values() []string {
	vals := make([]string, 0, 42)
	vals = append(vals, r.System, strconv.Itoa(int(r.Chart)))
	for _, v := range r.shield() {
		vals = append(vals, *v)
	}
	for _, v := range r.houseFigures() {
		vals = append(vals, *v)
	}
	for _, v := range r.houseSigns() {
		s := ""
		if *v != nil {
			s = **v
		}
		vals = append(vals, s)
	}
	return vals
}
