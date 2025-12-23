package handlers

import (
	"net/http"

	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func ActiveUTXOsByAddress(w http.ResponseWriter, r *http.Request) {
	logger := Log(r)
	db := DB(r)
	addressStr := chi.URLParam(r, "address")

	addr, _ := db.Address().GetByAddress(addressStr)
	if addr == nil {
		ape.RenderErr(w, problems.NotFound())
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
