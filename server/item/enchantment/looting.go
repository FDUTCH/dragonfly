package enchantment

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Looting is an enchantment that increases chance and amount of drop from mob.
var Looting looting

type looting struct {
}

func (l looting) Name() string {
	return "Looting"
}

func (l looting) MaxLevel() int {
	return 3
}

func (l looting) Cost(level int) (int, int) {
	min := 9*(level-1) + 15
	return min, 45 + min
}

func (l looting) Rarity() item.EnchantmentRarity {
	return item.EnchantmentRarityRare
}

func (l looting) CompatibleWithEnchantment(t item.EnchantmentType) bool {
	return true
}

func (l looting) CompatibleWithItem(i world.Item) bool {
	t, ok := i.(item.Tool)
	return ok && (t.ToolType() == item.TypeSword)
}
