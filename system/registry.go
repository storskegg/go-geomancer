package system

import (
	"fmt"
	"slices"
	"strings"
	"sync"
)

var (
	mu       sync.RWMutex
	registry = map[string]System{}
)

// Register makes a system available by its ID. It is meant to be called from
// init and panics if the system is invalid or its ID is already registered.
func Register(s System) {
	if err := Validate(s); err != nil {
		panic(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if _, dup := registry[s.ID()]; dup {
		panic(fmt.Sprintf("system: %q registered twice", s.ID()))
	}
	registry[s.ID()] = s
}

// Get returns the registered system with the given ID, ignoring case and
// surrounding space.
func Get(id string) (System, bool) {
	mu.RLock()
	defer mu.RUnlock()
	s, ok := registry[strings.ToLower(strings.TrimSpace(id))]
	return s, ok
}

// List returns the registered systems, sorted by ID.
func List() []System {
	mu.RLock()
	defer mu.RUnlock()
	ss := make([]System, 0, len(registry))
	for _, s := range registry {
		ss = append(ss, s)
	}
	slices.SortFunc(ss, func(a, b System) int { return strings.Compare(a.ID(), b.ID()) })
	return ss
}
