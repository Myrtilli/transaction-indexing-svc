package handlers

import (
	"net/http"

	"github.com/Myrtilli/transaction-indexing-svc/internal/service/models"
	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GetBalance(w http.ResponseWriter, r *http.Request) {
	logger := Log(r).WithField("handler", "GetBalance")
	db := DB(r)
	addressStr := chi.URLParam(r, "address")

	addr, err := db.Address().GetByAddress(addressStr)
	if err != nil {
		logger.WithError(err).Error("failed to get address")
		ape.RenderErr(w, problems.InternalError())
		return
	}
	if addr == nil {
		ape.RenderErr(w, problems.NotFound())
		return
	}

	lastBlock, err := db.BlockHeader().GetLast()
	if err != nil {
		logger.WithError(err).Error("failed to get last block header")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	var currentHeight int64
	if lastBlock != nil {
		currentHeight = lastBlock.Height
	}

	utxos, err := db.UTXO().SelectByAddressID(addr.ID)
	if err != nil {
		logger.WithError(err).Error("failed to select utxos")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	response := models.NewBalanceResponse(addressStr, utxos, currentHeight)

	logger.WithFields(map[string]interface{}{
		"address":   response.Address,
		"confirmed": response.ConfirmedBalance,
		"total":     response.TotalBalance,
	}).Info("balance calculated")

	ape.Render(w, response)
}
