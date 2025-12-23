package indexer

import "log"

func (i *Indexer) RollbackBlock(height int64) {
	actions := i.undoLog.Get(height)

	_ = i.db.NewTransaction(func() error {
		for _, a := range actions {
			switch a.Action {
			case "create_utxo":
				_ = i.db.UTXO().DeleteAboveHeight(height - 1)
			case "spend_utxo":
				_ = i.db.UTXO().UnspendByHeight(height - 1)
			}
		}
		return nil
	})

	i.undoLog.Remove(height)
	log.Printf("rolled back block %d", height)
}

func (i *Indexer) HandleReorg(newTipHeight int64) {
	commonAncestor := i.FindCommonAncestor(newTipHeight)
	for h := i.CurrentTip(); h > commonAncestor; h-- {
		i.RollbackBlock(h)
	}
}

func (i *Indexer) FindCommonAncestor(newHeight int64) int64 {
	for h := newHeight - 1; h > 0; h-- {
		dbBlock, _ := i.db.BlockHeader().GetByHeight(h)
		var rpcHash string
		i.rpcClient.Call("getblockhash", []any{h}, &rpcHash)

		if dbBlock != nil && dbBlock.BlockHash == rpcHash {
			return h
		}
		if newHeight-h > int64(i.cfg.MaxReorgDepth) {
			break
		}
	}
	return 0
}
