package loot

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"iter"
	"math/rand"
)

// Table represents loot table.
type Table struct {
	Pools []pool `json:"pools"`
}

// Loot returns an iterator that yields loot items.
func (t Table) Loot(seed int64) iter.Seq[item.Stack] {

	return func(yield func(item.Stack) bool) {
		r := rand.New(rand.NewSource(seed))

		for _, currentPool := range t.Pools {
			rollCount := count(r, currentPool.Rolls)
			for range rollCount {
				for stack := range choose(r, currentPool).stacks(r, seed, item.Stack{}, nil) {
					if !yield(stack) {
						return
					}
				}
			}
		}
	}

}

// FillInventory fills inventories, ensures inventory is filled without overlapping up to 100 slots.
func (t Table) FillInventory(r *rand.Rand, i *inventory.Inventory) (err error) {
	//defer func() {
	//	if e := recover(); e != nil {
	//		err = e.(error)
	//	}
	//}()

	nextSlot := fillAllSlots(i.Size(), r)

	for stack := range t.Loot(r.Int63()) {
		slot := nextSlot()
		err = i.SetItem(slot, stack)
		if err != nil {
			return err
		}
	}

	return
}

// fillAllSlots returns a closure which generates indices for filling inventory wise no overlaps.
func fillAllSlots(num int, r *rand.Rand) func() int {
	var itr = 0
	var current = 0

	for _, v := range primes {
		current = v
		if v > num+2 && num%v != 0 {
			itr = v
			break
		}
	}

	if itr == 0 {
		itr = current
	}

	val := itr + r.Intn(num)
	return func() int {
		val += itr
		return val % num
	}
}

// primes is an array of prime numbers used to generate fill inventory slots wise out overlapping.
var primes = [...]int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97}
