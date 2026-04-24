package world

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world/chunk"
)

type Executor struct {
	pos    Zone
	w      *World
	chunks [8][8]*Column

	// scheduledUpdates is a map of tick time values indexed by the block
	// position at which an update is scheduled. If the current tick exceeds the
	// tick value passed, the block update will be performed and the entry will
	// be removed from the map.
	scheduledUpdates *scheduledTickQueue
	entities         map[*EntityHandle]ChunkPos

	queue chan transaction
}

func (e *Executor) Pos() Zone {
	return e.pos
}

func (e *Executor) Block(pos cube.Pos) Block {
	withIn := e.posWithin(pos)
	ch := e.chunk(chunkPosFromPos(withIn))

}

func (e *Executor) chunk(pos ChunkPos) *Column {
	ch := e.chunks[pos.X()][pos.Z()]
	if ch != nil {
		return ch
	}
	e.w.loadChunk()
}

func (e *Executor) addChunk(pos ChunkPos, c *chunk.Column) *Column {
	column := e.columnFrom(c, pos)
	e.chunks[pos.X()][pos.Z()] = column
	for _, ent := range column.Entities {
		e.entities[ent] = pos
		ent.e = e
	}
	chunk.LightArea([]*chunk.Chunk{column.Chunk}, int(pos[0]), int(pos[1])).Fill()
	w.calculateLight(pos)
	return column
}

func (e *Executor) columnFrom(c *chunk.Column, _ ChunkPos) *Column {
	w := e.w
	col := &Column{
		Chunk:         c.Chunk,
		Entities:      make([]*EntityHandle, 0, len(c.Entities)),
		BlockEntities: make(map[cube.Pos]Block, len(c.BlockEntities)),
	}
	for _, e := range c.Entities {
		eid, ok := e.Data["identifier"].(string)
		if !ok {
			w.conf.Log.Error("read column: entity without identifier field", "ID", e.ID)
			continue
		}
		t, ok := w.conf.Entities.Lookup(eid)
		if !ok {
			w.conf.Log.Error("read column: unknown entity type", "ID", e.ID, "type", eid)
			continue
		}
		col.Entities = append(col.Entities, entityFromData(t, e.ID, e.Data))
	}
	for _, be := range c.BlockEntities {
		rid := c.Chunk.Block(uint8(be.Pos[0]), int16(be.Pos[1]), uint8(be.Pos[2]), 0)
		b, ok := BlockByRuntimeID(rid)
		if !ok {
			w.conf.Log.Error("read column: no block with runtime ID", "ID", rid)
			continue
		}
		nb, ok := b.(NBTer)
		if !ok {
			w.conf.Log.Error("read column: block with nbt does not implement NBTer", "block", fmt.Sprintf("%#v", b))
			continue
		}
		col.BlockEntities[be.Pos] = nb.DecodeNBT(be.Data).(Block)
	}
	scheduled, savedTick := make([]scheduledTick, 0, len(c.ScheduledBlocks)), c.Tick
	for _, t := range c.ScheduledBlocks {
		bl := blockByRuntimeIDOrAir(t.Block)
		scheduled = append(scheduled, scheduledTick{pos: t.Pos, b: bl, bhash: BlockHash(bl), t: e.scheduledUpdates.currentTick + (t.Tick - savedTick)})
	}
	e.scheduledUpdates.add(scheduled)
	return col
}

func (e *Executor) blockInChunk(c *Column, pos cube.Pos) Block {

}

func (e *Executor) posWithin(pos cube.Pos) cube.Pos {
	xOffset := int(e.pos.X() << 7)
	zOffset := int(e.pos.Z() << 7)
	return pos.Sub(cube.Pos{xOffset, 0, zOffset})
}

func chunkPosFromPos(pos cube.Pos) ChunkPos {
	return ChunkPos{
		int32(pos[0]) >> 4,
		int32(pos[2]) >> 4,
	}
}

func zonePosFromChunkPos(pos ChunkPos) Zone {
	return Zone{
		pos[0] >> 3,
		pos[1] >> 3,
	}
}

type Zone [2]int32

func (z Zone) X() int32 {
	return z[0]
}

func (z Zone) Z() int32 {
	return z[1]
}

func (z Zone) Within(pos cube.Pos) bool {
	const off = (1 << 7) - 1
	pos[1] = 0
	xMin := int(z.X() << 7)
	zMin := int(z.Z() << 7)
	xMax := xMin + off
	zMax := zMin + off
	return pos.Within(cube.Pos{xMin, 0, zMin}, cube.Pos{xMax, 0, zMax})
}
