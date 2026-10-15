package chart

import "testing"

func TestMatterLabel(t *testing.T) {
	husband := &Derivation{Person: 7, Subject: 10, PersonWord: "husband", SubjectWord: "job"}
	tests := []struct {
		m              *Matter
		subject, label string
	}{
		{nil, "", ""},
		{&Matter{}, "", ""},
		{&Matter{Topics: []string{"small animals"}}, "Small animals", "Small animals"},
		{&Matter{Topics: []string{"sickness", "small animals", "work"}}, "Sickness and small animals", "Sickness and small animals"},
		{&Matter{Topics: []string{"career"}, Derived: husband}, "Husband's career", "Husband's career (the 10th from the 7th)"},
		{&Matter{Topics: []string{"house"}, Derived: &Derivation{Person: 4, Subject: 4, PersonWord: "parents"}}, "Parents' house", "Parents' house (the 4th from the 4th)"},
		{&Matter{Topics: []string{"career"}, Derived: &Derivation{Person: 10, Subject: 10, PersonWord: "boss"}}, "Boss's career", "Boss's career (the 10th from the 10th)"},
	}
	for _, tt := range tests {
		if got := tt.m.Subject(); got != tt.subject {
			t.Errorf("%+v: Subject = %q, want %q", tt.m, got, tt.subject)
		}
		if got := tt.m.Label(); got != tt.label {
			t.Errorf("%+v: Label = %q, want %q", tt.m, got, tt.label)
		}
	}
}

func TestSetSignificatorClearsMatter(t *testing.T) {
	r := Reading{Significator: 6, Matter: &Matter{Topics: []string{"small animals"}}}
	if err := r.SetSignificator(7); err != nil {
		t.Fatal(err)
	}
	if r.Matter != nil {
		t.Errorf("Matter = %+v after changing the significator", r.Matter)
	}
}

func TestOrdinal(t *testing.T) {
	for n, want := range map[int]string{1: "1st", 2: "2nd", 3: "3rd", 4: "4th", 10: "10th", 11: "11th", 12: "12th", 21: "21st", 22: "22nd", 113: "113th"} {
		if got := Ordinal(n); got != want {
			t.Errorf("Ordinal(%d) = %q, want %q", n, got, want)
		}
	}
}
