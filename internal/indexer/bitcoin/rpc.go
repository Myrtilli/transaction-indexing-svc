package bitcoin

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type RPCClient struct {
	URL      string
	User     string
	Password string
	Client   *http.Client
}

func NewRPCClient(url, user, pass string) *RPCClient {
	return &RPCClient{
		URL:      url,
		User:     user,
		Password: pass,
		Client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *RPCClient) Call(method string, params []any, result any) error {
	payload := map[string]any{
		"jsonrpc": "1.0",
		"id":      "indexer",
		"method":  method,
		"params":  params,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", c.URL, bytes.NewReader(body))
	req.SetBasicAuth(c.User, c.Password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var rpcResp struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return err
	}
	if rpcResp.Error != nil {
		return errors.New(rpcResp.Error.Message)
	}
	return json.Unmarshal(rpcResp.Result, result)
}

func (c *RPCClient) GetBlockHeader(hash string) (*BlockHeader, error) {
	var header BlockHeader
	err := c.Call("getblockheader", []any{hash}, &header)
	return &header, err
}

func (c *RPCClient) GetBlock(hash string) ([]Transaction, error) {
	var block struct {
		Tx []Transaction `json:"tx"`
	}
	err := c.Call("getblock", []any{hash, 2}, &block)
	return block.Tx, err
}

func (c *RPCClient) GetTxOutProof(txid, blockhash string) ([]byte, error) {
	var proofHex string
	err := c.Call("gettxoutproof", []any{[]string{txid}, blockhash}, &proofHex)
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(proofHex)
}
