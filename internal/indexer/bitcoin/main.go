package bitcoin

import (
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

type Caller interface {
	Call(method string, params []any, result any) error
	GetBlockHash(height int64) (string, error)
	GetBlockHeader(hash string) (*BlockHeader, error)
	GetBlock(hash string) ([]Transaction, error)
	GetTxOutProof(txid, blockHash string) ([]byte, error)
}
