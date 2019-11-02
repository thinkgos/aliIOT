package dm

import (
	"sync"
)

type pool struct {
	pl sync.Pool
}

func newPool() *pool {
	return &pool{
		sync.Pool{
			New: func() interface{} { return &MsgCacheEntry{err: make(chan error, 1)} },
		},
	}
}

func (sf *pool) Get() *MsgCacheEntry {
	entry := sf.pl.Get().(*MsgCacheEntry)
loop:
	for {
		select {
		case <-entry.err:
		default:
			break loop
		}
	}

	return entry
}

func (sf *pool) Put(entry *MsgCacheEntry) {
	sf.pl.Put(entry)
}
