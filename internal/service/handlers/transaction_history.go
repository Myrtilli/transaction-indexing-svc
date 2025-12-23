package handlers

import (
	"net/http"

	"github.com/Myrtilli/transaction-indexing-svc/internal/service/models"
	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func TransactionHistoryByAddress(w http.ResponseWriter, r *http.Request) {
	logger := Log(r)
	db := DB(r)
	addressStr := chi.URLParam(r, "address")

	addr, err := db.Address().GetByAddress(addressStr)
	if err != nil {
		logger.WithError(err).Error("failed to get address from DB")
		ape.RenderErr(w, problems.InternalError())
		return
	}
	if addr == nil {
		ape.RenderErr(w, problems.NotFound())
		return
	}

	txs, err := db.Transaction().SelectByAddressID(addr.ID)
	if err != nil {
		logger.WithError(err).Error("failed to select transactions")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	lastBlock, _ := db.BlockHeader().GetLast()
	var height int64
	if lastBlock != nil {
		height = lastBlock.Height
	}

	logger.Infof("returned %d transactions for address %s", len(txs), addressStr)
	ape.Render(w, models.NewTxHistoryList(txs, height))
}
