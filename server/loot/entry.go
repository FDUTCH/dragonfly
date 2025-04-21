package loot

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/enchantment"
	"iter"
	"math/rand"
	"strings"
)

// entry represents loot entry.
type entry struct {
	Type      string `json:"type"`
	Weight    int    `json:"weight"`
	Functions []any  `json:"functions"`
	Name      string `json:"name"`
}

// stacks...
func (e entry) stacks(r *rand.Rand, seed int64, stack item.Stack, c KillСircumstances) iter.Seq[item.Stack] {
	lootingVal, _ := stack.Enchantment(enchantment.Looting)
	switch e.Type {
	case "loot_table":
		return Loot(e.Name, seed)
	case "empty":
		return empty
	case "item":
	default:
		panic(fmt.Errorf("unknown entry type %s", e.Type))
	}

	it, ok := itemByName(e.Name, 0)
	if !ok {
		return empty
	}

	var (
		meta         int16
		enchantments []item.Enchantment
		damage       float64
		minCount     int
		stackLore    []string
		customName   string
	)

	for _, fn := range e.Functions {
		function := fn.(map[string]any)
		fnType, _ := strings.CutPrefix(function["function"].(string), "minecraft:")
		switch fnType {
		case "set_count":
			if nbtconv.Bool(function, "add") {
				minCount += count(r, function["count"])
			} else {
				minCount = count(r, function["count"])
			}
		case "set_data":
			meta = nbtconv.Int16(function, "data")
		case "set_damage":
			if nbtconv.Bool(function, "add") {
				damage += itemDamage(r, function)
			} else {
				damage = itemDamage(r, function)
			}
		case "enchant_randomly":
			i, _ := itemByName(e.Name, meta)
			enchantments = append(enchantments, enchantRandomly(r, i, nbtconv.Bool(function, "treasure"))...)
		case "enchant_with_levels":
			i, _ := itemByName(e.Name, meta)
			enchantments = append(enchantments, enchantLevel(r, i, nbtconv.Bool(function, "treasure"), count(r, function["levels"]))...)
		case "specific_enchants":
			enchantments = append(enchantments, enchants(r, function)...)
		case "looting_enchant":
			minCount += r.Intn(lootingVal.Level() + 1)
		case "furnace_smelt":
			e.Name, meta = furnaceSmelt(it, c).EncodeItem()
		case "enchant_random_gear":
			chance := function["chance"].(float64)
			if useChance(chance, r) {
				i, _ := itemByName(e.Name, meta)
				enchantments = append(enchantments, enchantRandomly(r, i, nbtconv.Bool(function, "treasure"))...)
			}
		case "set_potion":
			meta = int16(potionId(function).Uint8())
		case "random_aux_value":
			meta = int16(count(r, function["values"]))
		case "set_stew_effect":
			meta = int16(stewEffect(r, function))
		case "set_lore":
			stackLore = lore(function)
		case "set_name":
			customName = nbtconv.String(function, "name")
		case "exploration_map", "random_block_state", "random_dye", "set_actor_id", "set_book_contents", "set_data_from_color_index", "fill_container", "trader_material_type":
			// can not be implemented currently.
		}
	}

	if minCount == 0 {
		minCount = 1
	}

	it, ok = itemByName(e.Name, meta)
	if !ok {
		return empty
	}
	result := item.NewStack(it, minCount)

	result = result.WithEnchantments(enchantments...).WithLore(stackLore...).WithCustomName(customName)

	result = result.WithDurability(int(float64(result.MaxDurability()) * (1 - damage)))
	return func(yield func(item.Stack) bool) {
		yield(result)
	}
}

// empty...
func empty(yield func(stack item.Stack) bool) {}
