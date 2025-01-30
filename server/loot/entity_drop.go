package loot

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"iter"
	"math/rand"
	"time"
)

type KillСircumstances interface {
	Entity() world.Entity
	Killer() world.Entity
}

func onFire(c KillСircumstances) bool {
	e, ok := c.Entity().(interface{ OnFireDuration() time.Duration })
	if !ok {
		return false
	}
	return e.OnFireDuration() > 0

}

func killedByPlayer(c KillСircumstances) bool {
	_, ok := c.Killer().(interface{ Food() int })
	return ok
}

func difficulty(c KillСircumstances) world.Difficulty {
	//TODO: get real difficulty
	return world.DifficultyHard
}

func itemHeld(c KillСircumstances) item.Stack {
	carrier, ok := c.Killer().(item.Carrier)
	if !ok {
		return item.Stack{}
	}
	main, _ := carrier.HeldItems()
	return main
}

func (t Table) EntityDrop(r *rand.Rand, c KillСircumstances) iter.Seq[item.Stack] {
	return func(yield func(item.Stack) bool) {
		seed := r.Int63()
		for _, currentPool := range t.Pools {
			if !currentPool.checkEntityDropConditions(r, c, itemHeld(c)) {
				continue
			}
			for range count(r, currentPool.Rolls) {
				for stack := range choose(r, currentPool).stacks(r, seed, itemHeld(c), c) {
					if !yield(stack) {
						return
					}
				}
			}
		}
	}
}
