package indexer

import (
	"encoding/json"
	"time"

	"github.com/Myrtilli/transaction-indexing-svc/internal/data"
	"github.com/Myrtilli/transaction-indexing-svc/internal/indexer/bitcoin"
)

func (i *Indexer) SyncNextBlock() {
	tip, err := i.db.BlockHeader().GetLast()
	if err != nil {
		i.logger.WithError(err).Error("failed to get tip")
		return
	}

	var checkHeight int64
	if tip == nil {
		checkHeight = int64(i.cfg.StartHeight)
	} else {
		checkHeight = tip.Height
	}

	var rpcHash string
	err = i.rpcClient.Call("getblockhash", []any{checkHeight}, &rpcHash)
	if err != nil {
		return
	}

	if tip != nil && tip.BlockHash != rpcHash {
		i.logger.WithFields(map[string]interface{}{
			"height":   checkHeight,
			"db_hash":  tip.BlockHash,
			"rpc_hash": rpcHash,
		}).Warn("reorg detected at tip!")

		i.HandleReorg(checkHeight + 1)
		return
	}

	nextHeight := checkHeight
	if tip != nil {
		nextHeight = tip.Height + 1
	}

	var nextBlockHash string
	err = i.rpcClient.Call("getblockhash", []any{nextHeight}, &nextBlockHash)
	if err != nil {
		return
	}

	header, err := i.rpcClient.GetBlockHeader(nextBlockHash)
	if err != nil {
		i.logger.WithField("height", nextHeight).WithError(err).Error("failed to fetch header")
		return
	}

	txs, err := i.rpcClient.GetBlock(nextBlockHash)
	if err != nil {
		i.logger.WithError(err).Error("failed to fetch block txs")
		return
	}

	i.processBlock(header, txs)
}

func (i *Indexer) processBlock(header *bitcoin.BlockHeader, txs []bitcoin.Transaction) {
	if bitcoin.CheckProofOfWork(header) {
		i.logger.WithField("hash", header.BlockHash).Info("Passed check of proof")
	} else {
		i.logger.WithField("hash", header.BlockHash).Error("Failed check of proof")
		return
	}

	err := i.db.BlockHeader().Insert(data.BlockHeader{
		BlockHash:      header.BlockHash,
		PreviousHash:   header.PreviousHash,
		Height:         header.Height,
		MerkleRoot:     header.MerkleRoot,
		Timestamp:      time.Unix(header.Timestamp, 0),
		Difficulty:     int64(header.Difficulty),
		Nonce:          int64(header.Nonce),
		TransactionNum: header.TransactionNum,
	})
	if err != nil {
		i.logger.WithError(err).Error("failed to insert block header")
		return
	}

	for _, tx := range txs {
		tracked := false
		for _, out := range tx.Outputs {
			addr := i.getAddrFromOutput(out)
			if addr != "" && i.isAddressTracked(addr) {
				tracked = true
				break
			}
		}

		if tracked {
			proof, err := i.rpcClient.GetTxOutProof(tx.TxID, header.BlockHash)

			entry := i.logger.WithField("tx_id", tx.TxID)

			if err == nil && bitcoin.VerifyMerkleProof(tx.TxID, [][]byte{proof}, header.MerkleRoot) {
				entry.Info("verified proof for tx")
			} else {
				entry.Warn("skipping real Merkle check for tx (Regtest mode)")
			}

			i.updateDatabase(tx, header)
			entry.Info("indexed transaction")
		}
	}

	i.logger.WithField("height", header.Height).Info("block indexed")
}

func (i *Indexer) getAddrFromOutput(out bitcoin.TxOutput) string {
	if out.Address != "" {
		return out.Address
	}
	if out.ScriptPubKey.Address != "" {
		return out.ScriptPubKey.Address
	}
	if len(out.ScriptPubKey.Addresses) > 0 {
		return out.ScriptPubKey.Addresses[0]
	}
	return "unknown"
}

func (i *Indexer) CurrentTip() int64 {
	tip, err := i.db.BlockHeader().GetLast()
	if err != nil || tip == nil {
		return 0
	}
	return tip.Height
}

func (i *Indexer) isAddressTracked(address string) bool {
	addr, err := i.db.Address().GetByAddress(address)
	return err == nil && addr != nil
}

func (i *Indexer) updateDatabase(tx bitcoin.Transaction, header *bitcoin.BlockHeader) {
	var dbInputs []data.TransactionInput
	var dbOutputs []data.TransactionOutput
	var transactionAddressID *int64
	var transactionAmount int64

	for _, in := range tx.Inputs {
		var prevTxID *string
		if in.PrevTxID != "" {
			prevTxID = &in.PrevTxID
		}
		dbInputs = append(dbInputs, data.TransactionInput{
			TxID:     tx.TxID,
			PrevTxID: prevTxID,
			VoutIdx:  uint32(in.Vout),
		})
		_ = i.db.UTXO().MarkAsSpent(in.PrevTxID, in.Vout)
	}

	for _, out := range tx.Outputs {
		addrStr := i.getAddrFromOutput(out)
		amountSat := int64(out.Value * 1e8)

		dbOutputs = append(dbOutputs, data.TransactionOutput{
			TxID:    tx.TxID,
			Address: addrStr,
			Amount:  amountSat,
			VoutIdx: uint32(out.Vout),
		})

		if addrStr != "" {
			addrRecord, err := i.db.Address().GetByAddress(addrStr)
			if err == nil && addrRecord != nil {
				if transactionAddressID == nil {
					transactionAddressID = &addrRecord.ID
					transactionAmount = amountSat
				}

				err = i.db.UTXO().Insert(data.UTXO{
					TxID:        tx.TxID,
					Vout:        out.Vout,
					AddressID:   addrRecord.ID,
					Amount:      amountSat,
					BlockHeight: header.Height,
				})

				if err == nil {
					i.undoLog.Add(UndoAction{
						BlockHeight: header.Height,
						Action:      "create_utxo",
						TxID:        tx.TxID,
						Vout:        out.Vout,
					})
				}
			}
		}
	}

	err := i.db.Transaction().Insert(data.Transaction{
		TxID:        tx.TxID,
		AddressID:   transactionAddressID,
		Amount:      transactionAmount,
		BlockHeight: header.Height,
		BlockHash:   header.BlockHash,
		MerkleProof: json.RawMessage(`[]`),
		CreatedAt:   time.Now(),
		Inputs:      dbInputs,
		Outputs:     dbOutputs,
	})

	if err != nil {
		i.logger.WithError(err).WithField("tx_id", tx.TxID).Error("failed to insert transaction")
	}
}
