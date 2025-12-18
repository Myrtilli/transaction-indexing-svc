package pg

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/Myrtilli/transaction-indexing-svc/internal/data"
	"gitlab.com/distributed_lab/kit/pgdb"
)

func newUserdb(db *pgdb.DB) data.Userdb {
	return &userU{
		db:  db,
		sql: sq.StatementBuilder,
	}
}

type userU struct {
	db  *pgdb.DB
	sql sq.StatementBuilderType
}

func (u *userU) Insert(user data.User) (*data.User, error) {
	query := sq.Insert("users").
		Columns("username", "hashed_password").
		Values(user.Username, user.PasswordHash).
		Suffix("RETURNING id")

	var id int64
	err := u.db.Get(&id, query)
	if err != nil {
		return nil, err
	}

	user.ID = id
	return &user, nil
}

func (u *userU) Get(username string) (*data.User, error) {
	query := sq.Select("*").
		From("users").
		Where(sq.Eq{"username": username})

	var user data.User
	err := u.db.Get(&user, query)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *userU) GetByUsername(username string) (*data.User, error) {
	query := sq.Select("*").
		From("users").
		Where(sq.Eq{"username": username})

	var user data.User
	err := u.db.Get(&user, query)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

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
