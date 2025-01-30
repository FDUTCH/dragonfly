package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/loot"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"math"
	"math/rand/v2"
	"slices"
)

const (
	// enchantingInputSlot is the slot index of the input item in the enchanting table.
	enchantingInputSlot = 0x0e
	// enchantingLapisSlot is the slot index of the lapis in the enchanting table.
	enchantingLapisSlot = 0x0f
)

// handleEnchant handles the enchantment of an item using the CraftRecipe stack request action.
func (h *ItemStackRequestHandler) handleEnchant(a *protocol.CraftRecipeStackRequestAction, s *Session, tx *world.Tx, c Controllable) error {
	// First ensure that the selected slot is not out of bounds.
	if a.RecipeNetworkID > 2 {
		return fmt.Errorf("invalid recipe network id: %d", a.RecipeNetworkID)
	}

	// Now ensure we have an input and only one input.
	input, err := s.ui.Item(enchantingInputSlot)
	if err != nil {
		return err
	}
	if input.Count() > 1 {
		return fmt.Errorf("enchanting tables only accept one item at a time")
	}

	// Determine the available enchantments using the session's enchantment seed.
	allCosts, allEnchants := s.determineAvailableEnchantments(tx, c, *s.openedPos.Load(), input)
	if len(allEnchants) == 0 {
		return fmt.Errorf("can't enchant non-enchantable item")
	}

	// Use the slot plus one as the cost. The requirement and enchantments can be found in the results from
	// determineAvailableEnchantments using the same slot index.
	cost := int(a.RecipeNetworkID + 1)
	requirement := allCosts[a.RecipeNetworkID]
	enchants := allEnchants[a.RecipeNetworkID]

	// If we don't have infinite resources, we need to deduct Lapis Lazuli and experience.
	if !c.GameMode().CreativeInventory() {
		// First ensure that the experience level is both underneath the requirement and the cost.
		if c.ExperienceLevel() < requirement {
			return fmt.Errorf("not enough levels to meet requirement")
		}
		if c.ExperienceLevel() < cost {
			return fmt.Errorf("not enough levels to meet cost")
		}

		// Then ensure that the player has input Lapis Lazuli, and enough of it to meet the cost.
		lapis, err := s.ui.Item(enchantingLapisSlot)
		if err != nil {
			return err
		}
		if _, ok := lapis.Item().(item.LapisLazuli); !ok {
			return fmt.Errorf("lapis lazuli was not input")
		}
		if lapis.Count() < cost {
			return fmt.Errorf("not enough lapis lazuli to meet cost")
		}

		// Deduct the experience and Lapis Lazuli.
		c.SetExperienceLevel(c.ExperienceLevel() - cost)
		h.setItemInSlot(protocol.StackRequestSlotInfo{
			Container: protocol.FullContainerName{ContainerID: protocol.ContainerEnchantingMaterial},
			Slot:      enchantingLapisSlot,
		}, lapis.Grow(-cost), s, tx)
	}

	// Reset the enchantment seed so different enchantments can be selected.
	c.ResetEnchantmentSeed()

	// Clear the existing input item, and apply the new item into the crafting result slot of the UI. The client will
	// automatically move the item into the input slot.
	h.setItemInSlot(protocol.StackRequestSlotInfo{
		Container: protocol.FullContainerName{ContainerID: protocol.ContainerEnchantingInput},
		Slot:      enchantingInputSlot,
	}, item.Stack{}, s, tx)

	return h.createResults(s, tx, input.WithEnchantments(enchants...))
}

// sendEnchantmentOptions sends a list of available enchantments to the client based on the client's enchantment seed
// and nearby bookshelves.
func (s *Session) sendEnchantmentOptions(tx *world.Tx, c Controllable, pos cube.Pos, stack item.Stack) {
	// First determine the available enchantments for the given item stack.
	selectedCosts, selectedEnchants := s.determineAvailableEnchantments(tx, c, pos, stack)
	if len(selectedEnchants) == 0 {
		// No available enchantments.
		return
	}

	// Build the protocol variant of the enchantment options.
	options := make([]protocol.EnchantmentOption, 0, 3)
	for i := 0; i < 3; i++ {
		// First build the enchantment instances for each selected enchantment.
		enchants := make([]protocol.EnchantmentInstance, 0, len(selectedEnchants[i]))
		for _, enchant := range selectedEnchants[i] {
			id, _ := item.EnchantmentID(enchant.Type())
			enchants = append(enchants, protocol.EnchantmentInstance{
				Type:  byte(id),
				Level: byte(enchant.Level()),
			})
		}

		// Then build the enchantment option. We can use the slot as the RecipeNetworkID, since the IDs seem to be unique
		// to enchanting tables only. We also only need to set the middle index of Enchantments. The other two serve
		// an unknown purpose and can cause various unexpected issues.
		options = append(options, protocol.EnchantmentOption{
			Name:            enchantNames[rand.IntN(len(enchantNames))],
			Cost:            uint32(selectedCosts[i]),
			RecipeNetworkID: uint32(i),
			Enchantments: protocol.ItemEnchantments{
				Slot:         int32(i),
				Enchantments: [3][]protocol.EnchantmentInstance{1: enchants},
			},
		})
	}

	// Send the enchantment options to the client.
	s.writePacket(&packet.PlayerEnchantOptions{Options: options})
}

// determineAvailableEnchantments returns a list of pseudo-random enchantments for the given item stack.
func (s *Session) determineAvailableEnchantments(tx *world.Tx, c Controllable, pos cube.Pos, stack item.Stack) ([]int, [][]item.Enchantment) {
	// First ensure that the item is enchantable and does not already have any enchantments.
	enchantable, ok := stack.Item().(item.Enchantable)
	if !ok {
		// We can't enchant this item.
		return nil, nil
	}
	if len(stack.Enchantments()) > 0 {
		// We can't enchant this item.
		return nil, nil
	}

	// Search for bookshelves around the enchanting table. Bookshelves help boost the value of the enchantments that
	// are selected, resulting in enchantments that are rarer but also more expensive.
	seed := uint64(c.EnchantmentSeed())
	random := rand.New(rand.NewPCG(seed, seed))
	bookshelves := searchBookshelves(tx, pos)
	value := enchantable.EnchantmentValue()

	// Calculate the base cost, used to calculate the upper, middle, and lower level costs.
	baseCost := random.IntN(8) + 1 + (bookshelves >> 1) + random.IntN(bookshelves+1)

	// Calculate the upper, middle, and lower level costs.
	upperLevelCost := max(baseCost/3, 1)
	middleLevelCost := baseCost*2/3 + 1
	lowerLevelCost := max(baseCost, bookshelves*2)

	// Create a list of available enchantments for each slot.
	return []int{
			upperLevelCost,
			middleLevelCost,
			lowerLevelCost,
		}, [][]item.Enchantment{
			loot.СreateEnchantments(random, stack.Item(), value, upperLevelCost, false),
			loot.СreateEnchantments(random, stack.Item(), value, middleLevelCost, false),
			loot.СreateEnchantments(random, stack.Item(), value, lowerLevelCost, false),
		}
}

// searchBookshelves searches for nearby bookshelves around the position passed, and returns the amount found.
func searchBookshelves(tx *world.Tx, pos cube.Pos) (shelves int) {
	for x := -1; x <= 1; x++ {
		for z := -1; z <= 1; z++ {
			for y := 0; y <= 1; y++ {
				if x == 0 && z == 0 {
					// Ignore the center block.
					continue
				}
				if _, ok := tx.Block(pos.Add(cube.Pos{x, y, z})).(block.Air); !ok {
					// There must be a one block space between the bookshelf and the player.
					continue
				}

				// Check for a bookshelf two blocks away.
				if _, ok := tx.Block(pos.Add(cube.Pos{x * 2, y, z * 2})).(block.Bookshelf); ok {
					shelves++
				}
				if x != 0 && z != 0 {
					// Check for a bookshelf two blocks away on the X axis.
					if _, ok := tx.Block(pos.Add(cube.Pos{x * 2, y, z})).(block.Bookshelf); ok {
						shelves++
					}
					// Check for a bookshelf two blocks away on the Z axis.
					if _, ok := tx.Block(pos.Add(cube.Pos{x, y, z * 2})).(block.Bookshelf); ok {
						shelves++
					}
				}

				if shelves >= 15 {
					// We've found enough bookshelves.
					return 15
				}
			}
		}
	}
	return shelves
}
