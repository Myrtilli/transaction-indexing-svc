package data

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
