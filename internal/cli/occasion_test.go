package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

var occasionNow = time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)

func TestReadOccasion(t *testing.T) {
	full := occasionFlags{date: "2026-10-15", time: "3:30 pm", place: "Springfield, IL"}
	tests := []struct {
		name        string
		f           occasionFlags
		interactive bool
		want        string // the resolved time, if all given
		need        asks
		err, warn   string
	}{
		{name: "none, without prompts", f: occasionFlags{}},
		{name: "none, with prompts", f: occasionFlags{}, interactive: true, need: asks{true, true, true}},
		{name: "all", f: full, want: "2026-10-15 15:30 CDT"},
		{name: "all, with prompts", f: full, interactive: true, want: "2026-10-15 15:30 CDT"},
		{name: "partial, without prompts", f: occasionFlags{date: "2026-10-15"}, err: "go together"},
		{name: "partial, with prompts", f: occasionFlags{date: "2026-10-15"}, interactive: true, need: asks{false, true, true}},
		{name: "bad date, without prompts", f: occasionFlags{date: "soon", time: "3pm", place: "Springfield, IL"}, err: `--date "soon"`},
		{name: "bad date, with prompts", f: occasionFlags{date: "soon"}, interactive: true, need: asks{true, true, true}, warn: `Couldn't use --date "soon"`},
		{name: "bad time, with prompts", f: occasionFlags{time: "teatime"}, interactive: true, need: asks{true, true, true}, warn: `--time "teatime"`},
		{name: "unknown place, with prompts", f: occasionFlags{place: "zzzzqq"}, interactive: true, need: asks{true, true, true}, warn: "no place by that name"},
		{name: "unknown place, without prompts", f: occasionFlags{date: "today", time: "noon", place: "zzzzqq"}, err: "no place by that name"},
		{
			name: "a time that happened twice, without prompts",
			f:    occasionFlags{date: "2026-11-01", time: "1:30 am", place: "Chicago, IL"},
			err:  "add CDT or CST",
		},
		{
			name: "a time that happened twice, with its zone",
			f:    occasionFlags{date: "2026-11-01", time: "1:30 am CST", place: "Chicago, IL"},
			want: "2026-11-01 01:30 CST",
		},
		{
			name: "a time that happened twice, with prompts", interactive: true,
			f:    occasionFlags{date: "2026-11-01", time: "1:30 am", place: "Chicago, IL"},
			want: "2026-11-01 01:30 CDT", // and the prompt asks which
		},
		{
			name: "a time that didn't occur, without prompts",
			f:    occasionFlags{date: "2026-03-08", time: "2:30 am", place: "Chicago, IL"},
			err:  "clocks went forward",
		},
		{
			name: "a time that didn't occur, with prompts", interactive: true,
			f:    occasionFlags{date: "2026-03-08", time: "2:30 am", place: "Chicago, IL"},
			need: asks{time: true}, warn: "clocks went forward",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var warn bytes.Buffer
			o, need, err := readOccasion(tt.f, tt.interactive, occasionNow, &warn)
			if tt.err != "" {
				if err == nil || !strings.Contains(err.Error(), tt.err) {
					t.Errorf("error = %v, want it to contain %q", err, tt.err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if need != tt.need {
				t.Errorf("asks = %+v, want %+v", need, tt.need)
			}
			if !strings.Contains(warn.String(), tt.warn) || (tt.warn == "" && warn.Len() > 0) {
				t.Errorf("warning = %q, want %q", warn.String(), tt.warn)
			}
			switch {
			case !tt.f.any() && !tt.interactive:
				if o != nil {
					t.Errorf("occasion = %+v, want none", o)
				}
			case tt.want != "":
				if got := o.Time().Format("2006-01-02 15:04 MST"); got != tt.want {
					t.Errorf("time = %s, want %s", got, tt.want)
				}
			}
		})
	}
}
