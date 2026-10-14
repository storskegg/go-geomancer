package chart

import "github.com/storskegg/go-geomancer/aspect"

// FocusHouse is a house that bears on the question, with its parts in it.
type FocusHouse struct {
	House        int
	Querent      bool          // House 1
	Significator bool          // the house of the question
	Aspect       *aspect.House // nil unless it aspects the significator
}

// Focus returns the houses that bear on the question, most important first:
// House 1 (the querent), the significator's house, then the houses aspecting
// it in the order of Reading.Aspects. A house with several parts, such as
// House 1 when it also aspects the significator, is listed once.
//
// Without a significator, every house is listed, House 1 first.
func (r Reading) Focus() []FocusHouse {
	if r.Significator == 0 {
		houses := make([]FocusHouse, 12)
		for i := range houses {
			houses[i] = FocusHouse{House: i + 1, Querent: i == 0}
		}
		return houses
	}

	houses := []FocusHouse{{House: 1, Querent: true}}
	at := map[int]int{1: 0} // house → index in houses
	add := func(h int) *FocusHouse {
		if i, ok := at[h]; ok {
			return &houses[i]
		}
		at[h] = len(houses)
		houses = append(houses, FocusHouse{House: h})
		return &houses[len(houses)-1]
	}
	add(r.Significator).Significator = true
	for _, a := range r.Aspects {
		add(a.House).Aspect = &a
	}
	return houses
}
