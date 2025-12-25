package indexer

func (i *Indexer) RollbackBlock(height int64) {
	actions := i.undoLog.Get(height)

	err := i.db.NewTransaction(func() error {
		for _, a := range actions {
			switch a.Action {
			case "create_utxo":
				_ = i.db.UTXO().DeleteAboveHeight(height - 1)
			case "spend_utxo":
				_ = i.db.UTXO().UnspendByHeight(height - 1)
			}
		}

		if err := i.db.BlockHeader().DeleteAboveHeight(height - 1); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		i.logger.WithError(err).Error("failed to rollback block header")
		return
	}

	i.undoLog.Remove(height)
	i.logger.WithField("height", height).Info("rolled back block and removed header")
}

func (i *Indexer) HandleReorg(newTipHeight int64) {
	i.logger.WithField("new_tip", newTipHeight).Info("reorganization detected, searching for common ancestor")

	commonAncestor := i.FindCommonAncestor(newTipHeight)
	currentTip := i.CurrentTip()

	i.logger.WithFields(map[string]interface{}{
		"common_ancestor": commonAncestor,
		"current_tip":     currentTip,
	}).Info("starting rollback process")

	for h := currentTip; h > commonAncestor; h-- {
		i.RollbackBlock(h)
	}
}

func (i *Indexer) FindCommonAncestor(newHeight int64) int64 {
	for h := newHeight - 1; h > 0; h-- {
		dbBlock, err := i.db.BlockHeader().GetByHeight(h)
		if err != nil {
			i.logger.WithError(err).WithField("height", h).Error("failed to get block from DB during reorg")
			continue
		}

		var rpcHash string
		err = i.rpcClient.Call("getblockhash", []interface{}{h}, &rpcHash)
		if err != nil {
			i.logger.WithError(err).WithField("height", h).Error("failed to get block hash from RPC")
			continue
		}

		if dbBlock != nil && dbBlock.BlockHash == rpcHash {
			i.logger.WithFields(map[string]interface{}{
				"height": h,
				"hash":   rpcHash,
			}).Debug("common ancestor found")
			return h
		}

		if newHeight-h > int64(i.cfg.MaxReorgDepth) {
			i.logger.Warn("max reorg depth reached, could not find common ancestor")
			break
		}
	}
	return 0
}
