package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Myrtilli/transaction-indexing-svc/internal/service/requests"
	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func ActiveUTXOsByAddress(w http.ResponseWriter, r *http.Request) {
	logger := Log(r)
	db := DB(r)
	addressStr := chi.URLParam(r, "address")
	userID := UserID(r)

	req, err := requests.Pagination(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	addresses := strings.Split(addressStr, ",")

	addr, err := db.Address().GetByAddressesUserID(addresses, userID)
	if err != nil {
		logger.WithError(err).Error("failed to get address from DB")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if len(addr) == 0 {
		err := errors.New(addressStr + " is not tracked, please, add them to your addresses list")
		logger.Error(err.Error())
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	var addressesIDs []int64
	for _, a := range addr {
		addressesIDs = append(addressesIDs, a.ID)
	}

	utxos, err := db.UTXO().SelectByAddressesID(addressesIDs, req.Pagination)
	if err != nil {
		logger.WithError(err).Error("failed to select utxos")
		ape.RenderErr(w, problems.InternalError())
		return
	}
	logger.Debugf("found %d active utxos for %s", len(utxos), addressStr)
	ape.Render(w, utxos)
}
