package models

import "github.com/Myrtilli/transaction-indexing-svc/internal/data"

type SuccessResponse struct {
	Message string `json:"message"`
}

const (
	RegistrationSuccessMessage = "User registered successfully"
	LoginSuccessMessage        = "User logged in successfully"
	NewAddressSuccessMessage   = "Address added successfully"
)

type AddressModel struct {
	ID      int64  `json:"id"`
	Address string `json:"address"`
}

func AddressList(src []data.Address) []AddressModel {
	res := make([]AddressModel, len(src))
	for i, v := range src {
		res[i] = AddressModel{
			ID:      v.ID,
			Address: v.Address,
		}
	}
	return res
}

type UTXOModel struct {
	TxID        string `json:"tx_id"`
	Vout        int    `json:"vout"`
	Amount      int64  `json:"amount"`
	BlockHeight int64  `json:"block_height"`
}

type TxHistoryItem struct {
	TxID          string            `json:"tx_id"`
	Amount        int64             `json:"amount"`
	BlockHeight   int64             `json:"block_height"`
	Confirmations int64             `json:"confirmations"`
	MerkleProof   []data.MerkleNode `json:"merkle_proof"`
	IsConfirmed   bool              `json:"is_confirmed"`
}

func NewTxHistoryList(txs []data.Transaction, currentHeight int64) []TxHistoryItem {
	res := make([]TxHistoryItem, len(txs))
	for i, tx := range txs {
		confirmations := currentHeight - tx.BlockHeight + 1
		if confirmations < 0 {
			confirmations = 0
		}

		var proof []data.MerkleNode
		res[i] = TxHistoryItem{
			TxID:          tx.TxID,
			Amount:        tx.Amount,
			BlockHeight:   tx.BlockHeight,
			Confirmations: confirmations,
			MerkleProof:   proof,
			IsConfirmed:   confirmations >= 6,
		}
	}
	return res
}

type BalanceResponse struct {
	Address            string `json:"address"`
	ConfirmedBalance   int64  `json:"confirmed_balance"`
	UnconfirmedBalance int64  `json:"unconfirmed_balance"`
	TotalBalance       int64  `json:"total_balance"`
}

func NewBalanceResponse(address string, utxos []data.UTXO, currentHeight int64) BalanceResponse {
	var confirmed, unconfirmed int64

	for _, u := range utxos {
		if currentHeight-u.BlockHeight >= 5 {
			confirmed += u.Amount
		} else {
			unconfirmed += u.Amount
		}
	}

	return BalanceResponse{
		Address:            address,
		ConfirmedBalance:   confirmed,
		UnconfirmedBalance: unconfirmed,
		TotalBalance:       confirmed + unconfirmed,
	}
}
