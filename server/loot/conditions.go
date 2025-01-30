package loot

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"math/rand"
)

func (p pool) checkEntityDropConditions(rand *rand.Rand, e KillСircumstances, stack item.Stack) (passed bool) {
	passed = true
	for _, condition := range p.Conditions {
		switch condition["condition"].(string) {
		case "killed_by_player_or_pets", "killed_by_player":
			passed = killedByPlayer(e)
		case "random_chance_with_looting":
			ench, _ := stack.Enchantment(enchantment.Looting)
			chance := condition["chance"].(float64)
			mul := condition["looting_multiplier"].(float64)
			mul *= float64(ench.Level())
			chance += mul

			passed = useChance(chance, rand)
		case "random_chance":
			chance := condition["chance"].(float64)
			if useChance(chance, rand) {
				continue
			}
			passed = false

		case "random_difficulty_chance":

			passed = false
		}

		if !passed {
			return
		}
	}
	return
}
