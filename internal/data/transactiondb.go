package data

import (
	"encoding/json"
	"time"
)

type Transactiondb interface {
	Insert(tx Transaction) error
	SelectByAddressesID(addressesIDs []int64, p Pagination, before *time.Time, after *time.Time) ([]Transaction, error)
	DeleteAboveHeight(height int64) error
}

type Pagination struct {
	Limit  uint64
	Offset uint64
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
	TimeStamp   int64
}

type TransactionInput struct {
	ID        int64   `db:"id"`
	TxID      string  `db:"tx_id"`
	PrevTxID  *string `db:"prev_tx_id"`
	VoutIdx   uint32  `db:"vout_idx"`
	Address   string  `db:"address"`
	Amount    int64   `db:"amount"`
	ScriptSig struct {
		Address   string   `json:"address"`
		Addresses []string `json:"addresses"`
	} `json:"scriptSig"`
}

type TransactionOutput struct {
	ID           int64  `db:"id"`
	TxID         string `db:"tx_id"`
	VoutIdx      uint32 `db:"vout_idx"`
	Address      string `db:"address"`
	Amount       int64  `db:"amount"`
	ScriptPubKey struct {
		Address   string   `json:"address"`
		Addresses []string `json:"addresses"`
	} `json:"scriptPubKey"`
}
