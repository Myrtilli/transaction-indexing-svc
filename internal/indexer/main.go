package indexer

import (
	"context"
	"time"

	"github.com/Myrtilli/transaction-indexing-svc/internal/data"
	"github.com/Myrtilli/transaction-indexing-svc/internal/indexer/bitcoin"
	"gitlab.com/distributed_lab/logan/v3"
)

type Config struct {
	MaxReorgDepth int
	PollInterval  time.Duration
}

type Indexer struct {
	db        data.MasterQ
	rpcClient *bitcoin.RPCClient
	cfg       Config
	logger    *logan.Entry
}

func New(logger *logan.Entry, db data.MasterQ, rpc *bitcoin.RPCClient, cfg Config) *Indexer {
	return &Indexer{
		logger:    logger.WithField("service", "indexer"),
		db:        db,
		rpcClient: rpc,
		cfg:       cfg,
	}
}

func (i *Indexer) Run(ctx context.Context) {
	i.logger.Info("indexer started")
	ticker := time.NewTicker(i.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			i.logger.Info("indexer stopped")
			return
		case <-ticker.C:
			i.SyncNextBlock()
		}
	}
}
