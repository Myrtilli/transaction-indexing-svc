package pg

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/Myrtilli/transaction-indexing-svc/internal/data"
	"gitlab.com/distributed_lab/kit/pgdb"
)

func newAddressdb(db *pgdb.DB) data.Addressdb {
	return &addressA{
		db:  db,
		sql: sq.StatementBuilder,
	}
}

type addressA struct {
	db  *pgdb.DB
	sql sq.StatementBuilderType
}

func (a *addressA) Insert(address data.Address) error {
	query := sq.Insert("addresses").
		Columns("user_id", "address").
		Values(address.UserID, address.Address)

	err := a.db.Exec(query)
	return err
}

func (a *addressA) Select(userID int64) ([]data.Address, error) {
	query := sq.Select("*").
		From("addresses").
		Where(sq.Eq{"user_id": userID})

	var addresses []data.Address
	err := a.db.Select(&addresses, query)
	if err != nil {
		return nil, err
	}

	return addresses, nil
}

func (a *addressA) GetByAddressUserID(address string, userID int64) (*data.Address, error) {
	var result data.Address
	query := sq.Select("*").From("addresses").Where(sq.Eq{"address": address, "user_id": userID})

	err := a.db.Get(&result, query)
	if err != nil {
		return nil, nil
	}
	return &result, err
}

func (a *addressA) GetByAddress(address string) (*data.Address, error) {
	query := sq.Select("*").
		From("addresses").
		Where(sq.Eq{"address": address})

	var addr data.Address
	err := a.db.Get(&addr, query)
	if err != nil {
		return nil, err
	}

	return &addr, nil
}

func (a *addressA) Get() (*data.Address, error) {
	var addr data.Address
	err := a.db.Get(&addr, a.sql.Select("*").From("addresses"))
	if err != nil {
		return nil, err
	}
	return &addr, nil
}
