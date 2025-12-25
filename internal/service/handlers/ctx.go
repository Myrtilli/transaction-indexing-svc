package handlers

import (
	"context"
	"net/http"

	"github.com/Myrtilli/transaction-indexing-svc/internal/config"
	"github.com/Myrtilli/transaction-indexing-svc/internal/data"
	"gitlab.com/distributed_lab/logan/v3"
)

type ctxKey int

const (
	logCtxKey      ctxKey = iota
	dbCtxKey       ctxKey = iota
	jwtCtxKey      ctxKey = iota
	usernameCtxKey ctxKey = iota
	indexerCtxKey  ctxKey = iota
	userIDCtxKey   ctxKey = iota
)

func CtxLog(entry *logan.Entry) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, logCtxKey, entry)
	}
}

func Log(r *http.Request) *logan.Entry {
	return r.Context().Value(logCtxKey).(*logan.Entry)
}

func CtxDB(entry data.MasterQ) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, dbCtxKey, entry)
	}
}

func DB(r *http.Request) data.MasterQ {
	return r.Context().Value(dbCtxKey).(data.MasterQ).New()
}

func CtxJWT(cfg config.Config) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, jwtCtxKey, cfg)
	}
}

func JWT(r *http.Request) string {
	cfg, ok := r.Context().Value(jwtCtxKey).(config.Config)
	if !ok {
		return ""
	}
	return cfg.JWTKey()
}

func Username(r *http.Request) string {
	val, ok := r.Context().Value(usernameCtxKey).(string)
	if !ok {
		return ""
	}
	return val
}

func UserID(r *http.Request) int64 {
	return r.Context().Value(userIDCtxKey).(int64)
}

func CtxIndexer(indexer interface{}) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, indexerCtxKey, indexer)
	}
}

func Indexer(r *http.Request) interface{} {
	return r.Context().Value(indexerCtxKey)
}
