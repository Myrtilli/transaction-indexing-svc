package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func ActiveUTXOsByAddress(w http.ResponseWriter, r *http.Request) {
	logger := Log(r)
	db := DB(r)
	addressStr := chi.URLParam(r, "address")
	userID := UserID(r)

	addr, _ := db.Address().GetByAddressUserID(addressStr, userID)
	if addr == nil {
		err := errors.New(addressStr + " is not tracked, please, add them to your addresses list")
		logger.Error(err.Error())
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	utxos, err := db.UTXO().SelectByAddressID(addr.ID)
	if err != nil {
		logger.WithError(err).Error("failed to select utxos")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	logger.Debugf("found %d active utxos for %s", len(utxos), addressStr)
	ape.Render(w, utxos)
}
