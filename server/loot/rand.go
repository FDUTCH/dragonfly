package loot

import "math/rand"

func choose(r *rand.Rand, p pool) entry {
	var all int
	for _, ent := range p.Entries {
		all += ent.Weight
	}
	if all == 0 {
		return entry{}
	}
	var value = r.Intn(all) + 1
	for _, ent := range p.Entries {
		value -= ent.Weight
		if value < 1 {
			return ent
		}
	}
	panic("un able to choose a random entry")
}
