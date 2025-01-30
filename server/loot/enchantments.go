package loot

import (
	"github.com/df-mc/dragonfly/server/internal/sliceutil"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"math"
	"math/rand"
	"slices"
)

// treasureEnchantment represents an enchantment that may be a treasure enchantment.
type treasureEnchantment interface {
	item.EnchantmentType
	Treasure() bool
}

// СreateEnchantments creates a list of enchantments for the given item stack and returns them.
func СreateEnchantments(random *rand.Rand, it world.Item, value, level int, treasure bool) []item.Enchantment {
	// Calculate the "random bonus" for this level. This factor is used in calculating the enchantment cost, used
	// during the selection of enchantments.
	randomBonus := (random.Float64() + random.Float64() - 1.0) * 0.15

	// Calculate the enchantment cost and clamp it to ensure it is always at least one with triangular distribution.
	cost := level + 1 + random.Intn(value/4+1) + random.Intn(value/4+1)
	cost = clamp(int(math.Round(float64(cost)+float64(cost)*randomBonus)), 1, math.MaxInt32)

	// Books are applicable to all enchantments, so make sure we have a flag for them here.

	_, book := it.(item.Book)

	// Now that we have our enchantment cost, we need to select the available enchantments. First, we iterate through
	// each possible enchantment.
	availableEnchants := make([]item.Enchantment, 0, len(item.Enchantments()))
	for _, enchant := range item.Enchantments() {
		if t, ok := enchant.(treasureEnchantment); ok && t.Treasure() && !treasure {
			// We then have to ensure that the enchantment is not a treasure enchantment, as those cannot be selected through
			// the enchanting table.
			continue
		}
		if !book && !enchant.CompatibleWithItem(it) {
			// The enchantment is not compatible with the item.
			continue
		}

		// Now iterate through each possible level of the enchantment.
		for i := enchant.MaxLevel(); i > 0; i-- {
			// Use the level to calculate the minimum and maximum costs for this enchantment.
			if minCost, maxCost := enchant.Cost(i); cost >= minCost && cost <= maxCost {
				// If the cost is within the bounds, add the enchantment to the list of available enchantments.
				availableEnchants = append(availableEnchants, item.NewEnchantment(enchant, i))
				break
			}
		}
	}
	if len(availableEnchants) == 0 {
		// No available enchantments, so we can't really do much here.
		return nil
	}

	// Now we need to select the enchantments.
	selectedEnchants := make([]item.Enchantment, 0, len(availableEnchants))

	// Select the first enchantment using a weighted random algorithm, favouring enchantments that have a higher weight.
	// These weights are based on the enchantment's rarity, with common and uncommon enchantments having a higher weight
	// than rare and very rare enchantments.
	enchant := weightedRandomEnchantment(random, availableEnchants)
	selectedEnchants = append(selectedEnchants, enchant)

	// Remove the selected enchantment from the list of available enchantments, so we don't select it again.
	ind := slices.Index(availableEnchants, enchant)
	availableEnchants = slices.Delete(availableEnchants, ind, ind+1)

	// Based on the cost, select a random amount of additional enchantments.
	for random.Intn(50) <= cost {
		// Ensure that we don't have any conflicting enchantments. If so, remove them from the list of available
		// enchantments.
		lastEnchant := selectedEnchants[len(selectedEnchants)-1]
		if availableEnchants = sliceutil.Filter(availableEnchants, func(enchant item.Enchantment) bool {
			return lastEnchant.Type().CompatibleWithEnchantment(enchant.Type())
		}); len(availableEnchants) == 0 {
			// We've exhausted all available enchantments.
			break
		}

		// Select another enchantment using the same weighted random algorithm.
		enchant = weightedRandomEnchantment(random, availableEnchants)
		selectedEnchants = append(selectedEnchants, enchant)

		// Remove the selected enchantment from the list of available enchantments, so we don't select it again.
		ind = slices.Index(availableEnchants, enchant)
		availableEnchants = slices.Delete(availableEnchants, ind, ind+1)

		// Halve the cost, so we have a lower chance of selecting another enchantment.
		cost /= 2
	}
	return selectedEnchants
}

// weightedRandomEnchantment returns a random enchantment from the given list of enchantments using the rarity weight of
// each enchantment.
func weightedRandomEnchantment(rs *rand.Rand, enchants []item.Enchantment) item.Enchantment {
	var totalWeight int
	for _, e := range enchants {
		totalWeight += e.Type().Rarity().Weight()
	}
	r := rs.Intn(totalWeight)
	for _, e := range enchants {
		r -= e.Type().Rarity().Weight()
		if r < 0 {
			return e
		}
	}
	panic("should never happen")
}

// clamp clamps a value into the given range.
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

var baseCost = 31

func upperLevelCost(baseCost int) int {
	return max(baseCost/3, 1)
}

func middleLevelCost(baseCost int) int {
	return baseCost*2/3 + 1
}

func lowerLevelCost(baseCost int) int {
	return max(baseCost, 15*2)
}

var enchantmentLevels = []func(int) int{
	upperLevelCost,
	middleLevelCost,
	lowerLevelCost,
}
