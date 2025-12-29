package data

import "time"

type BlockHeaderdb interface {
	Insert(BlockHeader) error
	GetByHeight(height int64) (*BlockHeader, error)
	GetByHash(hash string) (*BlockHeader, error)
	GetLast() (*BlockHeader, error)
	DeleteAboveHeight(height int64) error
}

type BlockHeader struct {
	BlockHash      string    `db:"block_hash"`
	PreviousHash   string    `db:"previous_hash"`
	TransactionNum int64     `db:"transaction_num"`
	Height         int64     `db:"height"`
	MerkleRoot     string    `db:"merkle_root"`
	Timestamp      time.Time `db:"timestamp"`
	Difficulty     int64     `db:"difficulty"`
	Nonce          int64     `db:"nonce"`
}
