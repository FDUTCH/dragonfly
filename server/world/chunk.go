package world

import (
	"errors"
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/internal/sliceutil"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/df-mc/goleveldb/leveldb"
	"sync"
)

type chunkManager struct {
	queue map[ChunkPos]*chunkRequest
	mu    sync.Mutex
	cl    *chunkLoader
}

func (c *chunkManager) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for pos, req := range c.queue {
		if req.closed() {
			delete(c.queue, pos)
		}
	}
}

func (c *chunkManager) request(pos ChunkPos, viewer Viewer) *chunkRequest {
	c.mu.Lock()
	defer c.mu.Unlock()
	tr, loaded := c.queue[pos]
	if !loaded {
		tr = newChunkRequest(pos, c.cl)
		c.queue[pos] = tr
	}
	if viewer != nil {
		tr.viewers.Append(viewer)
	}
	return tr
}

type chunkRequest struct {
	tx      chan struct{}
	ch      *Column
	err     error
	viewers *sliceutil.ThreadSafeSlice[Viewer]
}

func newChunkRequest(pos ChunkPos, source *chunkLoader) *chunkRequest {
	ct := &chunkRequest{tx: make(chan struct{}), viewers: &sliceutil.ThreadSafeSlice[Viewer]{}}
	go func() {
		ct.close(source.load(pos))
	}()
	return ct
}

func (c *chunkRequest) Chunk() (*Column, error) {
	<-c.tx
	return c.ch, c.err
}

func (c *chunkRequest) close(col *Column, err error) {
	c.ch = col
	c.err = err
	col.viewers = c.viewers
	close(c.tx)
}

func (c *chunkRequest) closed() bool {
	select {
	case <-c.tx:
		return true
	default:
		return false
	}
}

type chunkLoader struct {
	w *World
}

func (loader *chunkLoader) load(pos ChunkPos) (*Column, error) {
	conf := loader.w.conf
	ch, err := conf.Provider.LoadColumn(pos, loader.w.Dimension())
	switch {
	case err == nil:
		return loader.toColumn(ch, pos), nil
	case errors.Is(err, leveldb.ErrNotFound):
		c := chunk.New(airRID, loader.w.Range())
		conf.Generator.GenerateChunk(pos, c)
		ch = &chunk.Column{Chunk: c}
	default:
		ch = &chunk.Column{Chunk: chunk.New(airRID, loader.w.Range())}
	}
	return loader.toColumn(ch, pos), err
}

func (loader *chunkLoader) toColumn(c *chunk.Column, pos ChunkPos) *Column {
	col := &Column{
		Chunk:         c.Chunk,
		Entities:      make([]*EntityHandle, 0, len(c.Entities)),
		BlockEntities: make(map[cube.Pos]Block, len(c.BlockEntities)),
		updates:       c.ScheduledBlocks,
		tick:          c.Tick,
	}
	w := loader.w

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

	chunk.LightArea([]*chunk.Chunk{c.Chunk}, int(pos[0]), int(pos[1])).Fill()

	return col
}
