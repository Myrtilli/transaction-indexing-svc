package data

import (
	"encoding/json"
	"time"
)

type Transactiondb interface {
	Insert(tx Transaction) error
	SelectByAddressID(addressID int64) ([]Transaction, error)
	DeleteAboveHeight(height int64) error
}

type MerkleNode struct {
	Hash   string `json:"hash"`
	IsLeft bool   `json:"is_left"`
}

type Transaction struct {
	ID          int64               `db:"id"`
	TxID        string              `db:"tx_id"`
	AddressID   *int64              `db:"address_id"`
	Amount      int64               `db:"amount"`
	BlockHeight int64               `db:"block_height"`
	BlockHash   string              `db:"block_hash"`
	MerkleProof json.RawMessage     `db:"merkle_proof"`
	CreatedAt   time.Time           `db:"created_at"`
	Inputs      []TransactionInput  `db:"transaction_input"`
	Outputs     []TransactionOutput `db:"transaction_output"`
}

type TransactionInput struct {
	ID       int64   `db:"id"`
	TxID     string  `db:"tx_id"`
	PrevTxID *string `db:"prev_tx_id"`
	VoutIdx  uint32  `db:"vout_idx"`
	Address  string  `db:"address"`
	Amount   int64   `db:"amount"`
}

type TransactionOutput struct {
	ID      int64  `db:"id"`
	TxID    string `db:"tx_id"`
	Address string `db:"address"`
	Amount  int64  `db:"amount"`
	VoutIdx uint32 `db:"vout_idx"`
}
