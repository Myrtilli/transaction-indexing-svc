package pg

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/Myrtilli/transaction-indexing-svc/internal/data"
	"github.com/lib/pq"
	"gitlab.com/distributed_lab/kit/pgdb"
)

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

	if err := t.db.Exec(query); err != nil {
		return err
	}

	for _, input := range tx.Inputs {
		input_query := sq.Insert("transaction_inputs").
			Columns("tx_id", "address", "amount", "vout_idx").
			Values(tx.TxID, input.Address, input.Amount, input.VoutIdx)

		if err := t.db.Exec(input_query); err != nil {
			return err
		}
	}

	for _, output := range tx.Outputs {
		output_query := sq.Insert("transaction_outputs").
			Columns("tx_id", "address", "amount", "vout_idx").
			Values(tx.TxID, output.Address, output.Amount, output.VoutIdx)

		if err := t.db.Exec(output_query); err != nil {
			return err
		}
	}

	return nil
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

	for i := range transactions {
		inputs, err := t.getInputsForTransaction(transactions[i].TxID)
		if err != nil {
			return nil, err
		}
		transactions[i].Inputs = inputs

		outputs, err := t.getOutputsForTransaction(transactions[i].TxID)
		if err != nil {
			return nil, err
		}
		transactions[i].Outputs = outputs
	}

	return transactions, nil
}

func (t *transactionT) DeleteAboveHeight(height int64) error {
	query := sq.Delete("transactions").
		Where(sq.Gt{"block_height": height})

	err := t.db.Exec(query)
	return err
}

func (t *transactionT) getInputsForTransaction(txID string) ([]data.TransactionInput, error) {
	var inputs []data.TransactionInput
	query := sq.Select("*").From("transaction_inputs").Where(sq.Eq{"tx_id": txID})
	err := t.db.Select(&inputs, query)
	return inputs, err
}

func (t *transactionT) getOutputsForTransaction(txID string) ([]data.TransactionOutput, error) {
	var outputs []data.TransactionOutput
	query := sq.Select("*").From("transaction_outputs").Where(sq.Eq{"tx_id": txID})
	err := t.db.Select(&outputs, query)
	return outputs, err
}
