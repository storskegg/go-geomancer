package export

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "modernc.org/sqlite" // registers the "sqlite" driver

	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/position"
	"github.com/storskegg/go-geomancer/system"
	"github.com/storskegg/go-geomancer/zodiac"
)

// schema creates the tables. Figures and signs are stored by integer ID
// (the figure's 4-bit value; 0 = Aries) and described in lookup tables.
// charts holds each Shield once, independent of system; chart_houses holds
// each system's Square.
const schema = `
CREATE TABLE figures (
	id        INTEGER PRIMARY KEY CHECK (id BETWEEN 0 AND 15),
	name      TEXT    NOT NULL UNIQUE,
	pattern   TEXT    NOT NULL UNIQUE,
	planet    TEXT    NOT NULL,
	dots      INTEGER NOT NULL,
	odd_dots  INTEGER NOT NULL,
	even_dots INTEGER NOT NULL
);

CREATE TABLE signs (
	id   INTEGER PRIMARY KEY CHECK (id BETWEEN 0 AND 11),
	name TEXT    NOT NULL UNIQUE
);

CREATE TABLE positions (
	code      TEXT    PRIMARY KEY,
	ordinal   INTEGER NOT NULL UNIQUE,
	principal INTEGER NOT NULL CHECK (principal IN (0, 1))
);

CREATE TABLE systems (
	id           TEXT PRIMARY KEY,
	name         TEXT NOT NULL,
	head_element TEXT NOT NULL,
	neck_element TEXT NOT NULL,
	body_element TEXT NOT NULL,
	feet_element TEXT NOT NULL
);

CREATE TABLE system_figures (
	system_id TEXT    NOT NULL REFERENCES systems (id),
	figure_id INTEGER NOT NULL REFERENCES figures (id),
	sign_id   INTEGER REFERENCES signs (id),
	element   TEXT,
	PRIMARY KEY (system_id, figure_id)
);

CREATE TABLE system_houses (
	system_id TEXT    NOT NULL REFERENCES systems (id),
	position  TEXT    NOT NULL REFERENCES positions (code),
	house     INTEGER NOT NULL CHECK (house BETWEEN 1 AND 12),
	PRIMARY KEY (system_id, position),
	UNIQUE (system_id, house)
);

CREATE TABLE charts (
	id INTEGER PRIMARY KEY CHECK (id BETWEEN 0 AND 65535),
	m1 INTEGER NOT NULL REFERENCES figures (id),
	m2 INTEGER NOT NULL REFERENCES figures (id),
	m3 INTEGER NOT NULL REFERENCES figures (id),
	m4 INTEGER NOT NULL REFERENCES figures (id),
	d1 INTEGER NOT NULL REFERENCES figures (id),
	d2 INTEGER NOT NULL REFERENCES figures (id),
	d3 INTEGER NOT NULL REFERENCES figures (id),
	d4 INTEGER NOT NULL REFERENCES figures (id),
	n1 INTEGER NOT NULL REFERENCES figures (id),
	n2 INTEGER NOT NULL REFERENCES figures (id),
	n3 INTEGER NOT NULL REFERENCES figures (id),
	n4 INTEGER NOT NULL REFERENCES figures (id),
	wr INTEGER NOT NULL REFERENCES figures (id),
	wl INTEGER NOT NULL REFERENCES figures (id),
	j  INTEGER NOT NULL REFERENCES figures (id),
	jm INTEGER NOT NULL REFERENCES figures (id)
);

CREATE TABLE chart_houses (
	system_id TEXT    NOT NULL REFERENCES systems (id),
	chart_id  INTEGER NOT NULL REFERENCES charts (id),
	house     INTEGER NOT NULL CHECK (house BETWEEN 1 AND 12),
	position  TEXT    NOT NULL REFERENCES positions (code),
	figure_id INTEGER NOT NULL REFERENCES figures (id),
	sign_id   INTEGER REFERENCES signs (id),
	PRIMARY KEY (system_id, chart_id, house)
) WITHOUT ROWID;
`

// indexes are created after loading, which is faster than maintaining them
// row by row.
const indexes = `
CREATE INDEX chart_houses_by_figure ON chart_houses (system_id, house, figure_id);
CREATE INDEX charts_by_judge ON charts (j);
`

// sqliteWriter loads records into a new SQLite database in one transaction.
// Lookup tables are filled up front; each system's own tables when its first
// record arrives.
type sqliteWriter struct {
	ctx     context.Context
	db      *sql.DB
	tx      *sql.Tx
	chart   *sql.Stmt
	house   *sql.Stmt
	systems map[string]bool
	charts  []bool // indexed by chart ID: already in charts
}

// NewSQLiteWriter creates the database at path, which must not exist.
func NewSQLiteWriter(ctx context.Context, path string) (Writer, error) {
	// Bulk-loading a new file: no journal is needed, and a failed export
	// leaves a file the caller discards. A larger page cache (256 MiB) and
	// in-memory temporary storage speed up the load and the index sorts.
	db, err := sql.Open("sqlite", "file:"+path+"?mode=rwc"+
		"&_pragma=journal_mode(OFF)&_pragma=synchronous(OFF)&_pragma=foreign_keys(ON)"+
		"&_pragma=cache_size(-262144)&_pragma=temp_store(MEMORY)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	w := &sqliteWriter{ctx: ctx, db: db, systems: map[string]bool{}, charts: make([]bool, 1<<16)}
	if err := w.init(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite: %w", err)
	}
	return w, nil
}

func (w *sqliteWriter) init() error {
	var n int
	if err := w.db.QueryRowContext(w.ctx, "SELECT count(*) FROM sqlite_master").Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return errors.New("database is not empty")
	}
	if _, err := w.db.ExecContext(w.ctx, schema); err != nil {
		return err
	}

	tx, err := w.db.BeginTx(w.ctx, nil)
	if err != nil {
		return err
	}
	w.tx = tx
	if err := w.insertLookups(); err != nil {
		return err
	}
	if w.chart, err = tx.PrepareContext(w.ctx, insertSQL("charts", 17)); err != nil {
		return err
	}
	// One statement inserts a whole Square, which is much faster than a
	// statement per house.
	w.house, err = tx.PrepareContext(w.ctx, insertRowsSQL("chart_houses", 6, 12))
	return err
}

func insertSQL(table string, cols int) string { return insertRowsSQL(table, cols, 1) }

func insertRowsSQL(table string, cols, rows int) string {
	row := "(" + strings.TrimSuffix(strings.Repeat("?,", cols), ",") + ")"
	return "INSERT INTO " + table + " VALUES " + strings.TrimSuffix(strings.Repeat(row+",", rows), ",")
}

func (w *sqliteWriter) exec(query string, args ...any) error {
	_, err := w.tx.ExecContext(w.ctx, query, args...)
	return err
}

func (w *sqliteWriter) insertLookups() error {
	for _, f := range figure.All() {
		if err := w.exec(insertSQL("figures", 7),
			int(f), f.Name(), f.Pattern(), f.Planet().String(), f.Dots(), f.OddDots(), f.EvenDots()); err != nil {
			return err
		}
	}
	for _, s := range zodiac.Signs() {
		if err := w.exec(insertSQL("signs", 2), int(s), s.String()); err != nil {
			return err
		}
	}
	for _, p := range position.All() {
		if err := w.exec(insertSQL("positions", 3), p.String(), int(p), p.Principal()); err != nil {
			return err
		}
	}
	return nil
}

func (w *sqliteWriter) insertSystem(s system.System) error {
	args := []any{s.ID(), s.Name()}
	for _, seg := range figure.Segments() {
		args = append(args, s.SegmentElement(seg).String())
	}
	if err := w.exec(insertSQL("systems", 6), args...); err != nil {
		return err
	}
	for _, f := range figure.All() {
		var sign, elem any
		if v, ok := s.Sign(f); ok {
			sign = int(v)
		}
		if v, ok := s.Element(f); ok {
			elem = v.String()
		}
		if err := w.exec(insertSQL("system_figures", 4), s.ID(), int(f), sign, elem); err != nil {
			return err
		}
	}
	for _, p := range position.Principals() {
		h, _ := s.House(p)
		if err := w.exec(insertSQL("system_houses", 3), s.ID(), p.String(), h); err != nil {
			return err
		}
	}
	return nil
}

func (w *sqliteWriter) Write(rec Record) error {
	if id := rec.System.ID(); !w.systems[id] {
		if err := w.insertSystem(rec.System); err != nil {
			return fmt.Errorf("sqlite: system %s: %w", id, err)
		}
		w.systems[id] = true
	}

	if !w.charts[rec.Chart] {
		args := make([]any, 0, 17)
		args = append(args, rec.Chart)
		for _, p := range position.All() {
			args = append(args, int(rec.Reading.Shield.At(p)))
		}
		if _, err := w.chart.ExecContext(w.ctx, args...); err != nil {
			return fmt.Errorf("sqlite: chart %d: %w", rec.Chart, err)
		}
		w.charts[rec.Chart] = true
	}

	args := make([]any, 0, 6*len(rec.Reading.Square))
	for _, h := range rec.Reading.Square {
		var sign any
		if h.Sign != nil {
			sign = int(*h.Sign)
		}
		args = append(args, rec.System.ID(), rec.Chart, h.Number, h.Position.String(), int(h.Figure), sign)
	}
	if _, err := w.house.ExecContext(w.ctx, args...); err != nil {
		return fmt.Errorf("sqlite: chart %d houses: %w", rec.Chart, err)
	}
	return nil
}

// Close commits the load, builds the indexes and closes the database.
func (w *sqliteWriter) Close() error {
	err := w.tx.Commit()
	if err == nil {
		_, err = w.db.ExecContext(w.ctx, indexes)
	}
	if cerr := w.db.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return fmt.Errorf("sqlite: %w", err)
	}
	return nil
}
