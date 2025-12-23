package pg

import (
	"github.com/Myrtilli/transaction-indexing-svc/internal/data"
	"gitlab.com/distributed_lab/kit/pgdb"
)

func NewMasterQ(db *pgdb.DB) data.MasterQ {
	return &masterQ{
		db: db.Clone(),
	}
}

type masterQ struct {
	db *pgdb.DB
}

func (m *masterQ) New() data.MasterQ {
	return NewMasterQ(m.db)
}

func (m *masterQ) User() data.Userdb {
	return newUserdb(m.db)
}

func (m *masterQ) Address() data.Addressdb {
	return newAddressdb(m.db)
}

func (m *masterQ) BlockHeader() data.BlockHeaderdb {
	return newBlockHeaderdb(m.db)
}

func (m *masterQ) Transaction() data.Transactiondb {
	return newTransactiondb(m.db)
}

func (m *masterQ) UTXO() data.UTXOdb {
	return newUTXOdb(m.db)
}

func (m *masterQ) NewTransaction(fn func() error) error {
	return m.db.Transaction(func() error {
		return fn()
	})
}
