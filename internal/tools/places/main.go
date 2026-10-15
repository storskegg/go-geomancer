// Command places builds place/data/places.tsv.gz, the US places the place
// package searches, from GeoNames. It's run by hand when the data changes,
// not as part of the build; its output is committed.
//
// GeoNames data is licensed under CC BY 4.0 (https://www.geonames.org).
//
//	curl -LO https://download.geonames.org/export/dump/US.zip && unzip US.zip US.txt
//	curl -LO https://download.geonames.org/export/dump/admin2Codes.txt
//	go run ./internal/tools/places US.txt admin2Codes.txt place/data/places.tsv.gz
//
// It keeps populated places (feature class P) in the 50 states and DC,
// leaving out sections of places (PPLX) and abandoned, historical and
// destroyed ones (PPLQ, PPLH, PPLW). Where GeoNames lists the same name in
// the same county more than once, it keeps the most populous.
//
// The output is gzipped, tab-separated text, most populous first: name,
// ASCII name, state, county, latitude, longitude, population, and GeoNames'
// time zone (used only to check the time zone lookup).
package main

import (
	"bufio"
	"cmp"
	"compress/gzip"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

var states = strings.Fields(`AL AK AZ AR CA CO CT DE DC FL GA HI ID IL IN IA KS KY LA ME MD MA MI MN MS
MO MT NE NV NH NJ NM NY NC ND OH OK OR PA RI SC SD TN TX UT VT VA WA WV WI WY`)

var excluded = map[string]bool{"PPLX": true, "PPLQ": true, "PPLH": true, "PPLW": true}

type row struct {
	name, ascii, state, county, lat, lon, tz string
	pop                                      int
}

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: places US.txt admin2Codes.txt OUTPUT.tsv.gz")
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2], os.Args[3]); err != nil {
		fmt.Fprintln(os.Stderr, "places:", err)
		os.Exit(1)
	}
}

func lines(path string, fn func([]string) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<24)
	for sc.Scan() {
		if err := fn(strings.Split(sc.Text(), "\t")); err != nil {
			return err
		}
	}
	return sc.Err()
}

func run(placesPath, countiesPath, out string) error {
	counties := map[string]string{} // "IL.167" → "Sangamon County"
	err := lines(countiesPath, func(f []string) error {
		if code, ok := strings.CutPrefix(f[0], "US."); ok && len(f) > 1 {
			counties[code] = f[1]
		}
		return nil
	})
	if err != nil {
		return err
	}

	wanted := map[string]bool{}
	for _, s := range states {
		wanted[s] = true
	}
	best := map[string]row{} // ASCII name, state, county → most populous
	err = lines(placesPath, func(f []string) error {
		if len(f) < 19 || f[6] != "P" || excluded[f[7]] || !wanted[f[10]] {
			return nil
		}
		pop, _ := strconv.Atoi(f[14])
		r := row{name: f[1], ascii: f[2], state: f[10], county: counties[f[10]+"."+f[11]], lat: f[4], lon: f[5], pop: pop, tz: f[17]}
		key := strings.ToLower(r.ascii) + "\t" + r.state + "\t" + r.county
		if prev, ok := best[key]; !ok || r.pop > prev.pop {
			best[key] = r
		}
		return nil
	})
	if err != nil {
		return err
	}

	rows := make([]row, 0, len(best))
	for _, r := range best {
		rows = append(rows, r)
	}
	slices.SortFunc(rows, func(a, b row) int {
		return cmp.Or(cmp.Compare(b.pop, a.pop), strings.Compare(a.ascii, b.ascii), strings.Compare(a.state, b.state), strings.Compare(a.county, b.county))
	})

	o, err := os.Create(out)
	if err != nil {
		return err
	}
	gz, err := gzip.NewWriterLevel(o, gzip.BestCompression)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(gz)
	for _, r := range rows {
		// bufio.Writer keeps the first error, which Flush returns.
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\n", r.name, r.ascii, r.state, r.county, r.lat, r.lon, r.pop, r.tz)
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	fmt.Printf("wrote %d places to %s\n", len(rows), out)
	return o.Close()
}
