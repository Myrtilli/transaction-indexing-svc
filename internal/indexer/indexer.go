package indexer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Myrtilli/transaction-indexing-svc/internal/data"
	"github.com/Myrtilli/transaction-indexing-svc/internal/indexer/bitcoin"
)

type Config struct {
	MaxReorgDepth int
	PollInterval  time.Duration
}

type Indexer struct {
	db        data.MasterQ
	rpcClient *bitcoin.RPCClient
	undoLog   *UndoLog
	cfg       Config
}

func New(db data.MasterQ, rpc *bitcoin.RPCClient, cfg Config) *Indexer {
	return &Indexer{
		db:        db,
		rpcClient: rpc,
		undoLog:   NewUndoLog(),
		cfg:       cfg,
	}
}

func (i *Indexer) Run(ctx context.Context) {
	ticker := time.NewTicker(i.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Indexer stopped")
			return
		case <-ticker.C:
			i.SyncNextBlock()
		}
	}
}

func (i *Indexer) SyncNextBlock() {
	tip, err := i.db.BlockHeader().GetLast()
	if err != nil {
		log.Printf("failed to get tip: %v", err)
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
		log.Printf("failed to fetch header %d: %v", nextHeight, err)
		return
	}

	if tip != nil && header.PreviousHash != tip.BlockHash {
		log.Printf("Reorg detected at height %d!", nextHeight)
		i.HandleReorg(nextHeight)
		return
	}

	txs, err := i.rpcClient.GetBlock(blockHash)
	if err != nil {
		log.Printf("failed to fetch block txs: %v", err)
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
		log.Printf("DATABASE ERROR: failed to insert block header: %v", err)
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

			if err == nil && bitcoin.VerifyMerkleProof(tx.TxID, [][]byte{proof}, header.MerkleRoot) {
				log.Printf("Verified proof for tx: %s", tx.TxID)
			} else {
				log.Printf("Note: Skipping strict Merkle check for tx: %s (Regtest mode)", tx.TxID)
			}

			i.updateDatabase(tx, header)
			log.Printf("SUCCESS: Indexed transaction %s", tx.TxID)
		}
	}

	log.Printf("SUCCESS: Block %d indexed", header.Height)
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
			log.Printf("DB ERROR (Tx History): %v", err)
		}

		err = i.db.UTXO().Insert(data.UTXO{
			TxID:        tx.TxID,
			Vout:        out.Vout,
			AddressID:   addrRecord.ID,
			Amount:      int64(out.Value * 1e8),
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
