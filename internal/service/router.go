package service

import (
	"github.com/Myrtilli/transaction-indexing-svc/internal/config"
	"github.com/Myrtilli/transaction-indexing-svc/internal/data/pg"
	"github.com/Myrtilli/transaction-indexing-svc/internal/service/handlers"
	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"
)

func (s *service) router(cfg config.Config) chi.Router {
	r := chi.NewRouter()

	r.Use(
		ape.RecoverMiddleware(s.log),
		ape.LoganMiddleware(s.log),
		ape.CtxMiddleware(
			handlers.CtxLog(s.log),
			handlers.CtxDB(pg.NewMasterQ(cfg.DB())),
			handlers.CtxJWT(cfg),
		),
	)

	r.Route("/integrations/transaction-indexing-svc", func(r chi.Router) {
		r.Post("/login", handlers.Login)
		r.Post("/register", handlers.Register)

		r.Route("/", func(r chi.Router) {
			r.Use(handlers.AuthRequired)
			r.Post("/addresses", handlers.NewAddress)
		})
	})

	return r
}
