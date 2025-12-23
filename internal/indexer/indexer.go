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

	nextHeight := int64(0)
	if tip != nil {
		nextHeight = tip.Height + 1
	}

	var blockHash string
	err = i.rpcClient.Call("getblockhash", []any{nextHeight}, &blockHash)
	if err != nil {
		return
	}

	header, err := i.rpcClient.GetBlockHeader(blockHash)
	if err != nil {
		i.logger.WithField("height", nextHeight).WithError(err).Error("failed to fetch header")
		return
	}

	txs, err := i.rpcClient.GetBlock(blockHash)
	if err != nil {
		i.logger.WithError(err).Error("failed to fetch block txs")
		return
	}

	i.processBlock(header, txs)
}

func (i *Indexer) processBlock(header *bitcoin.BlockHeader, txs []bitcoin.Transaction) {
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
				entry.Warn("skipping strict Merkle check for tx (Regtest mode)")
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
	return ""
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
	for _, in := range tx.Inputs {
		_ = i.db.UTXO().MarkAsSpent(in.PrevTxID, in.Vout)
	}

	for _, out := range tx.Outputs {
		addrStr := i.getAddrFromOutput(out)
		addrRecord, err := i.db.Address().GetByAddress(addrStr)
		if err != nil || addrRecord == nil {
			continue
		}

		err = i.db.Transaction().Insert(data.Transaction{
			TxID:        tx.TxID,
			AddressID:   addrRecord.ID,
			Amount:      int64(out.Value * 1e8),
			BlockHeight: header.Height,
			BlockHash:   header.BlockHash,
			MerkleProof: json.RawMessage(`[]`),
			CreatedAt:   time.Now(),
		})
		if err != nil {
			i.logger.WithError(err).WithField("tx_id", tx.TxID).Error("failed to insert tx history")
		}

		err = i.db.UTXO().Insert(data.UTXO{
			TxID:        tx.TxID,
			Vout:        out.Vout,
			AddressID:   addrRecord.ID,
			Amount:      int64(out.Value * 1e8),
			BlockHeight: header.Height,
		})
		if err != nil {
			i.logger.WithError(err).WithField("tx_id", tx.TxID).Error("failed to insert utxo")
		}
	}
}
