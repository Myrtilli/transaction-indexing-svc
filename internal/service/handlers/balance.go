package handlers

import (
	"errors"
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
	userID := UserID(r)

	username, ok := r.Context().Value(usernameCtxKey).(string)
	if !ok {
		logger.Error("username not found in context")
		ape.RenderErr(w, problems.Unauthorized())
		return
	}

	userRecord, err := db.User().GetByUsername(username)
	if err != nil {
		logger.WithError(err).Error("failed to get user from db")
		ape.RenderErr(w, problems.InternalError())
		return
	}
	if userRecord == nil {
		ape.RenderErr(w, problems.Unauthorized())
		return
	}

	addr, err := db.Address().GetByAddressUserID(addressStr, userID)
	if err != nil {
		logger.WithError(err).Error("failed to get address")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if addr == nil {
		err := errors.New(addressStr + " is not tracked, please, add them to your addresses list")
		logger.Error(err.Error())
		ape.RenderErr(w, problems.BadRequest(err)...)
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
