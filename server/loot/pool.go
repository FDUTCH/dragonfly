package loot

import "math/rand"

// pool represents loot pool.
type pool struct {
	Rolls      any              `json:"rolls"`
	Entries    []entry          `json:"entries"`
	Conditions []map[string]any `json:"conditions"`
}

func (p pool) rolls(r *rand.Rand) int {
	return count(r, p.Rolls)
}
