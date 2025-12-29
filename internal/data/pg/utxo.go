package pg

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/Myrtilli/transaction-indexing-svc/internal/data"
	"gitlab.com/distributed_lab/kit/pgdb"
)

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
