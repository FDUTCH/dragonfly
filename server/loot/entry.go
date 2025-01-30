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

func (e entry) stacks(r *rand.Rand, seed int64, stack item.Stack, c KillСircumstances) iter.Seq[item.Stack] {

	lootingVal, _ := stack.Enchantment(enchantment.Looting)
	switch e.Type {
	case "loot_table":
		return Loot(e.Name, seed)
	case "empty":
		return stump
	case "item":
	default:
		panic(fmt.Errorf("unknown entry type %s", e.Type))
	}
	it, ok := itemByName(e.Name, 0)
	if !ok {
		return stump
	}
	var (
		data         int16
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
			data = nbtconv.Int16(function, "data")
		case "set_damage":
			if nbtconv.Bool(function, "add") {
				damage += itemDamage(r, function)
			} else {
				damage = itemDamage(r, function)
			}
		case "enchant_randomly":
			i, _ := itemByName(e.Name, data)
			enchantments = append(enchantments, enchantRandomly(r, i, nbtconv.Bool(function, "treasure"))...)
		case "enchant_with_levels":
			i, _ := itemByName(e.Name, data)
			enchantments = append(enchantments, enchantLevel(r, i, nbtconv.Bool(function, "treasure"), count(r, function["levels"]))...)
		case "specific_enchants":
			enchantments = append(enchantments, enchants(r, function)...)
		case "looting_enchant":
			minCount += r.Intn(lootingVal.Level() + 1)
		case "furnace_smelt":
			e.Name, data = furnaceSmelt(it, c).EncodeItem()
		case "enchant_random_gear":
			chance := function["chance"].(float64)
			if useChance(chance, r) {
				i, _ := itemByName(e.Name, data)
				enchantments = append(enchantments, enchantRandomly(r, i, nbtconv.Bool(function, "treasure"))...)
			}
		case "set_potion":
			data = int16(potionId(function).Uint8())
		case "random_aux_value":
			data = int16(count(r, function["values"]))
		case "set_stew_effect":
			data = int16(stewEffect(r, function))
		case "set_lore":
			stackLore = lore(function)
		case "set_name":
			customName = nbtconv.String(function, "name")
		case "exploration_map", "random_block_state", "random_dye", "set_actor_id", "set_book_contents", "set_data_from_color_index", "fill_container", "trader_material_type":
			//
		}
	}

	if minCount == 0 {
		minCount = 1
	}

	it, ok = itemByName(e.Name, data)
	if !ok {
		return stump
	}
	result := item.NewStack(it, minCount)

	result = result.WithEnchantments(enchantments...).WithLore(stackLore...).WithCustomName(customName)

	result = result.WithDurability(int(float64(result.MaxDurability()) * (1 - damage)))
	return func(yield func(item.Stack) bool) {
		yield(result)
	}
}

func stump(yield func(stack item.Stack) bool) {}
