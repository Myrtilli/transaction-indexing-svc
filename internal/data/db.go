package data

import (
	"encoding/json"
	"time"
)

type Userdb interface {
	Insert(User) (*User, error)
	Get(username string) (*User, error)
	GetByUsername(username string) (*User, error)
}

type User struct {
	ID           int64  `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"hashed_password"`
}

type Addressdb interface {
	Insert(Address) error
	Select(userID int64) ([]Address, error)
	GetByAddress(address string) (*Address, error)
	Get() (*Address, error)
	GetByAddressUserID(address string, userID int64) (*Address, error)
}

type Address struct {
	ID      int64  `db:"id"`
	UserID  int64  `db:"user_id"`
	Address string `db:"address"`
}

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
	ID          int64           `db:"id"`
	TxID        string          `db:"tx_id"`
	AddressID   int64           `db:"address_id"`
	Amount      int64           `db:"amount"`
	BlockHeight int64           `db:"block_height"`
	BlockHash   string          `db:"block_hash"`
	MerkleProof json.RawMessage `db:"merkle_proof"`
	CreatedAt   time.Time       `db:"created_at"`
}

type UTXOdb interface {
	Insert(utxo UTXO) error
	SelectByAddressID(addressID int64) ([]UTXO, error)
	MarkAsSpent(txID string, vout int64) error
	DeleteAboveHeight(height int64) error
	UnspendByHeight(height int64) error
	FilterByHeight(height int64) UTXOdb
	Get() (*UTXO, error)
}

type UTXO struct {
	ID          int64  `db:"id"`
	AddressID   int64  `db:"address_id"`
	TxID        string `db:"tx_id"`
	Vout        int64  `db:"vout"`
	Amount      int64  `db:"amount"`
	BlockHeight int64  `db:"block_height"`
	IsSpent     bool   `db:"is_spent"`
}
