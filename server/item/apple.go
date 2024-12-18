package item

import (
	"github.com/df-mc/dragonfly/server/world"
)

// Apple is a food item that can be eaten by the player.
type Apple struct {
	defaultFood
}

// Consume ...
func (a Apple) Consume(_ *world.Tx, c Consumer) Stack {
	c.Saturate(4, 2.4)
	return Stack{}
}

// CompostChance ...
func (Apple) CompostChance() float64 {
	return 0.65
}

// EncodeItem ...
func (a Apple) EncodeItem() (name string, meta int16) {
	return "minecraft:apple", 0
}
