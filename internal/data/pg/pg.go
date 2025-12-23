package pg

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/Myrtilli/transaction-indexing-svc/internal/data"
	"github.com/lib/pq"
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

func (a *addressA) FilterByAddress(address string) data.Addressdb {
	a.sql = a.sql.Where(sq.Eq{"address": address})
	return a
}

func (a *addressA) Get() (*data.Address, error) {
	var addr data.Address
	err := a.db.Get(&addr, a.sql.Select("*").From("addresses"))
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

func newBlockHeaderdb(db *pgdb.DB) data.BlockHeaderdb {
	return &blockHeaderB{
		db:  db,
		sql: sq.StatementBuilder,
	}
}

type blockHeaderB struct {
	db  *pgdb.DB
	sql sq.StatementBuilderType
}

func (b *blockHeaderB) Insert(header data.BlockHeader) error {
	query := sq.Insert("block_headers").
		Columns("block_hash", "previous_hash", "transaction_num", "height", "merkle_root", "timestamp", "difficulty", "nonce").
		Values(header.BlockHash, header.PreviousHash, header.TransactionNum, header.Height, header.MerkleRoot, header.Timestamp, header.Difficulty, header.Nonce).
		PlaceholderFormat(sq.Dollar)

	err := b.db.Exec(query)
	return err
}

func (b *blockHeaderB) GetByHeight(height int64) (*data.BlockHeader, error) {
	query := sq.Select("*").
		From("block_headers").
		Where(sq.Eq{"height": height})

	var header data.BlockHeader
	err := b.db.Get(&header, query)
	if err != nil {
		return nil, err
	}

	return &header, nil
}

func (b *blockHeaderB) GetLast() (*data.BlockHeader, error) {
	query := sq.Select(
		"block_hash",
		"previous_hash",
		"transaction_num",
		"height",
		"merkle_root",
		"timestamp",
		"difficulty",
		"nonce",
	).
		From("block_headers").
		OrderBy("height DESC").
		Limit(1).
		PlaceholderFormat(sq.Dollar)

	var header data.BlockHeader

	err := b.db.Get(&header, query)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	return &header, nil
}

func (b *blockHeaderB) DeleteAboveHeight(height int64) error {
	query := sq.Delete("block_headers").
		Where(sq.Gt{"height": height})

	err := b.db.Exec(query)
	return err
}

func (b *blockHeaderB) GetByHash(hash string) (*data.BlockHeader, error) {
	query := sq.Select("*").
		From("block_headers").
		Where(sq.Eq{"block_hash": hash})

	var header data.BlockHeader
	err := b.db.Get(&header, query)
	if err != nil {
		return nil, err
	}

	return &header, nil
}

func newTransactiondb(db *pgdb.DB) data.Transactiondb {
	return &transactionT{
		db:  db,
		sql: sq.StatementBuilder,
	}
}

type transactionT struct {
	db  *pgdb.DB
	sql sq.StatementBuilderType
}

func (t *transactionT) Insert(tx data.Transaction) error {
	query := sq.Insert("transactions").
		Columns("tx_id", "address_id", "amount", "block_height", "block_hash", "merkle_proof").
		Values(tx.TxID, tx.AddressID, tx.Amount, tx.BlockHeight, tx.BlockHash, pq.Array(tx.MerkleProof))

	err := t.db.Exec(query)
	return err
}

func (t *transactionT) SelectByAddressID(addressID int64) ([]data.Transaction, error) {
	query := sq.Select("*").
		From("transactions").
		Where(sq.Eq{"address_id": addressID})

	var transactions []data.Transaction
	err := t.db.Select(&transactions, query)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

func (t *transactionT) DeleteAboveHeight(height int64) error {
	query := sq.Delete("transactions").
		Where(sq.Gt{"block_height": height})

	err := t.db.Exec(query)
	return err
}

func newUTXOdb(db *pgdb.DB) data.UTXOdb {
	return &utxoU{
		db:  db,
		sql: sq.StatementBuilder,
	}
}

type utxoU struct {
	db  *pgdb.DB
	sql sq.StatementBuilderType
}

func (u *utxoU) Insert(utxo data.UTXO) error {
	query := sq.Insert("utxos").
		Columns("address_id", "tx_id", "vout", "amount", "block_height", "is_spent").
		Values(utxo.AddressID, utxo.TxID, utxo.Vout, utxo.Amount, utxo.BlockHeight, utxo.IsSpent)

	err := u.db.Exec(query)
	return err
}

func (u *utxoU) SelectByAddressID(addressID int64) ([]data.UTXO, error) {
	query := sq.Select("*").
		From("utxos").
		Where(sq.Eq{"address_id": addressID})

	var utxos []data.UTXO
	err := u.db.Select(&utxos, query)
	if err != nil {
		return nil, err
	}

	return utxos, nil
}

func (u *utxoU) MarkAsSpent(txID string, vout int64) error {
	query := sq.Update("utxos").
		Set("is_spent", true).
		Where(sq.Eq{"tx_id": txID, "vout": vout})

	err := u.db.Exec(query)
	return err
}

func (u *utxoU) DeleteAboveHeight(height int64) error {
	query := sq.Delete("utxos").
		Where(sq.Gt{"block_height": height})

	err := u.db.Exec(query)
	return err
}

func (u *utxoU) UnspendByHeight(height int64) error {
	query := sq.Update("utxos").
		Set("is_spent", false).
		Where(sq.Eq{"block_height": height})

	err := u.db.Exec(query)
	return err
}

func (u *utxoU) FilterByHeight(height int64) data.UTXOdb {
	u.sql = u.sql.Where(sq.Eq{"block_height": height})
	return u
}

func (u *utxoU) Get() (*data.UTXO, error) {
	{
		var utxo data.UTXO
		err := u.db.Get(&utxo, u.sql.Select("*").From("utxos"))
		if err != nil {
			return nil, err
		}
		return &utxo, nil
	}
}
