package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Myrtilli/transaction-indexing-svc/internal/service/models"
	"github.com/Myrtilli/transaction-indexing-svc/internal/service/requests"
	"github.com/go-chi/chi"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func TransactionHistoryByAddress(w http.ResponseWriter, r *http.Request) {
	logger := Log(r)
	db := DB(r)
	addressStr := chi.URLParam(r, "address")
	userID := UserID(r)

	req, err := requests.Pagination(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	time_req, err := requests.TimeFilter(r)
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

	txs, err := db.Transaction().SelectByAddressesID(addressesIDs, req.Pagination, time_req.Before, time_req.After)
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
