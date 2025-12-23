package data

type MasterQ interface {
	New() MasterQ
	User() Userdb
	Address() Addressdb
	BlockHeader() BlockHeaderdb
	Transaction() Transactiondb
	UTXO() UTXOdb
	NewTransaction(fn func() error) error
}
