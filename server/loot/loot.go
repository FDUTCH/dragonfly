package loot

import (
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/df-mc/dragonfly/server/internal/nbtconv"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
	"io/fs"
	"iter"
	"math/rand"
	"sync"
)

//go:embed loot_tables
var loots embed.FS

// Loot finds or creates new loot.Table and returns an iterator that yields items from this loot.Table.
func Loot(path string, seed int64) iter.Seq[item.Stack] {
	table := NewTable(path)
	return table.Loot(seed)
}

// NewTable finds or creates new loot.Table.
func NewTable(path string) Table {
	mu.RLock()
	loot, ok := mp[path]
	mu.RUnlock()
	var err error
	if !ok {
		var file fs.File
		file, err = loots.Open(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		dec := json.NewDecoder(file)
		err = dec.Decode(&loot)

		mu.Lock()
		mp[path] = loot
		mu.Unlock()
	}
	return loot
}

var (
	mu sync.RWMutex
	mp = map[string]Table{}
)

func count(r *rand.Rand, val any) int {
	var result int
	switch a := val.(type) {
	case float32:
		result = int(a)
	case float64:
		result = int(a)
	case int:
		result = a
	case map[string]any:
		max := unknownToInteger(a["max"])
		min := unknownToInteger(a["min"])
		result = r.Intn(max-min) + min
	default:
		panic(fmt.Sprintf("unknown type %#v", val))
	}
	return result
}

func itemDamage(r *rand.Rand, val map[string]any) float64 {
	m := val["damage"].(map[string]any)
	Max := int(m["max"].(float64) * 1000)
	Min := int(m["min"].(float64) * 1000)
	return float64(r.Intn(Max-Min)+Min) / 1000
}

func enchantRandomly(r *rand.Rand, it world.Item, treasure bool) []item.Enchantment {
	index := r.Intn(3)
	fn := enchantmentLevels[index]
	val, ok := it.(item.Enchantable)
	if !ok {
		return nil
	}
	return СreateEnchantments(r, it, val.EnchantmentValue(), fn(baseCost), treasure)
}

func enchantLevel(r *rand.Rand, it world.Item, treasure bool, level int) []item.Enchantment {
	val, ok := it.(item.Enchantable)
	if !ok {
		return nil
	}
	return СreateEnchantments(r, it, val.EnchantmentValue(), level, treasure)
}

func unknownToInteger(val any) int {
	switch v := val.(type) {
	case float32:
		return int(v)
	case float64:
		return int(v)
	case int:
		return v
	case int32:
		return int(v)
	default:
		panic(fmt.Sprintf("non number type : %T", val))
	}
}

func enchants(r *rand.Rand, val map[string]any) (result []item.Enchantment) {
	enchants := val["enchants"].([]any)
	for _, ench := range enchants {
		m := ench.(map[string]any)
		id := nbtconv.String(m, "id")
		level := m["level"].([]any)
		e, ok := enchantmentStringIds[id]

		if !ok {
			continue
		}

		t, registered := item.EnchantmentByID(e)
		if registered {
			result = append(result, item.NewEnchantment(t, int(level[r.Intn(len(level))].(float64))))
		}

	}
	return result
}

func furnaceSmelt(it world.Item, c KillСircumstances) world.Item {
	if onFire(c) {
		if smeltable, ok := it.(item.Smeltable); ok {
			it = smeltable.SmeltInfo().Product.Item()
		}
	}
	return it
}

func useChance(chance float64, r *rand.Rand) bool {
	val := r.Intn(100)
	return float64(val)/100 <= chance
}

//func difficultyChance(val map[string]any, difficulty world.Difficulty) bool {
//
//	chance := "default_chance"
//
//	switch difficulty {
//	case world.DifficultyHard:
//
//	case world.DifficultyNormal:
//	case world.DifficultyEasy:
//	case world.DifficultyPeaceful:
//
//	}
//}

func stewEffect(r *rand.Rand, val map[string]any) int {
	effects := val["effects"].([]any)
	return unknownToInteger(effects[r.Intn(len(effects))])
}

func potionId(val map[string]any) potion.Potion {
	return potionNames[val["id"].(string)]
}

func lore(val map[string]any) []string {
	arr := nbtconv.Slice(val, "lore")
	l := make([]string, 0, len(arr))
	for _, v := range arr {
		l = append(l, v.(string))
	}
	return l
}
