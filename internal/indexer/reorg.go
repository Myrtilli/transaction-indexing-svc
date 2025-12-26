package indexer

func (i *Indexer) RollbackBlock(height int64) {
	actions := i.undoLog.Get(height)

	err := i.db.NewTransaction(func() error {
		for _, a := range actions {
			switch a.Action {
			case "create_utxo":
				if err := i.db.UTXO().DeleteAboveHeight(height - 1); err != nil {
					return err
				}
			case "spend_utxo":
				if err := i.db.UTXO().UnspendByHeight(height - 1); err != nil {
					return err
				}
			}
		}

		if err := i.db.BlockHeader().DeleteAboveHeight(height - 1); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		i.logger.WithError(err).WithField("height", height).Error("failed to rollback block")
		return
	}

	i.undoLog.Remove(height)
	i.logger.WithField("height", height).Info("rolled back block and removed header")
}

func (i *Indexer) HandleReorg(newTipHeight int64) {
	commonAncestor := i.FindCommonAncestor(newTipHeight)
	currentTip := i.CurrentTip()

	i.logger.WithFields(map[string]interface{}{
		"common_ancestor": commonAncestor,
		"current_tip":     currentTip,
		"new_tip":         newTipHeight,
	}).Info("starting reorganization process")

	for h := currentTip; h > commonAncestor; h-- {
		i.RollbackBlock(h)
	}
}

func (i *Indexer) FindCommonAncestor(newHeight int64) int64 {
	currentTip := i.CurrentTip()

	for h := currentTip; h > 0; h-- {
		if currentTip-h > int64(i.cfg.MaxReorgDepth) {
			i.logger.WithFields(map[string]interface{}{
				"max_depth": i.cfg.MaxReorgDepth,
				"current_h": h,
			}).Error("max reorg depth reached")
			break
		}

		dbBlock, err := i.db.BlockHeader().GetByHeight(h)
		if err != nil || dbBlock == nil {
			continue
		}

		var rpcHash string
		err = i.rpcClient.Call("getblockhash", []interface{}{h}, &rpcHash)
		if err != nil {
			i.logger.WithError(err).WithField("height", h).Error("rpc failure during ancestor search")
			continue
		}

		if dbBlock.BlockHash == rpcHash {
			return h
		}
	}
	return 0
}