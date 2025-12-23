package indexer

import (
	"sync"
)

type UndoAction struct {
	BlockHeight int64
	Action      string
	TxID        string
	Vout        int64
}

type UndoLog struct {
	mu   sync.Mutex
	logs map[int64][]UndoAction
}

func NewUndoLog() *UndoLog {
	return &UndoLog{
		logs: make(map[int64][]UndoAction),
	}
}

func (u *UndoLog) Add(action UndoAction) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.logs[action.BlockHeight] = append(u.logs[action.BlockHeight], action)
}

func (u *UndoLog) Get(blockHeight int64) []UndoAction {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.logs[blockHeight]
}

func (u *UndoLog) Remove(blockHeight int64) {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.logs, blockHeight)
}
