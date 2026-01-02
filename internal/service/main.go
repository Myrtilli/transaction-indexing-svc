package service

import (
	"context"
	"net"
	"net/http"

	"github.com/Myrtilli/transaction-indexing-svc/internal/config"
	"github.com/Myrtilli/transaction-indexing-svc/internal/data/pg"
	"github.com/Myrtilli/transaction-indexing-svc/internal/indexer"
	"github.com/Myrtilli/transaction-indexing-svc/internal/indexer/bitcoin"
	"gitlab.com/distributed_lab/kit/copus/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type service struct {
	log      *logan.Entry
	copus    types.Copus
	listener net.Listener
	indexer  *indexer.Indexer
}

func (s *service) run(cfg config.Config) error {
	s.log.Info("Service started")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		s.log.Info("Starting background indexer loop")
		s.indexer.Run(ctx)
	}()

	r := s.router(cfg)

	if err := s.copus.RegisterChi(r); err != nil {
		return errors.Wrap(err, "cop failed")
	}

	return http.Serve(s.listener, r)
}

func newService(cfg config.Config) *service {
	db := pg.NewMasterQ(cfg.DB())
	rpc := bitcoin.NewRPCClient(cfg.NodeURL(), cfg.NodeUser(), cfg.NodePass())

	idx := indexer.New(cfg.Log(), db, rpc, indexer.Config{
		MaxReorgDepth: 6,
		PollInterval:  cfg.IndexerPollInterval(),
		StartHeight:   int(cfg.StartHeight()),
	})

	return &service{
		log:      cfg.Log(),
		copus:    cfg.Copus(),
		listener: cfg.Listener(),
		indexer:  idx,
	}
}

func Run(cfg config.Config) {
	if err := newService(cfg).run(cfg); err != nil {
		panic(err)
	}
}
