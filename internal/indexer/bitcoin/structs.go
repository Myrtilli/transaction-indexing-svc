package bitcoin

type BlockHeader struct {
	BlockHash      string  `json:"hash"`
	PreviousHash   string  `json:"previousblockhash"`
	MerkleRoot     string  `json:"merkleroot"`
	Timestamp      int64   `json:"time"`
	Bits           string  `json:"bits"`
	Nonce          uint32  `json:"nonce"`
	Height         int64   `json:"height"`
	TransactionNum int64   `json:"nTx"`
	Difficulty     float64 `json:"difficulty"`
}

type Transaction struct {
	TxID    string     `json:"txid"`
	Inputs  []TxInput  `json:"vin"`
	Outputs []TxOutput `json:"vout"`
}

type TxInput struct {
	PrevTxID string `json:"txid"`
	Vout     int64  `json:"vout"`
}

type TxOutput struct {
	Value        float64 `json:"value"`
	Vout         int64   `json:"n"`
	ScriptPubKey struct {
		Address   string   `json:"address"`
		Addresses []string `json:"addresses"`
	} `json:"scriptPubKey"`
	Address string `json:"address,omitempty"`
}
