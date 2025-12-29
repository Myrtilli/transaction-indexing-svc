package pg

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/Myrtilli/transaction-indexing-svc/internal/data"
	"gitlab.com/distributed_lab/kit/pgdb"
)

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
