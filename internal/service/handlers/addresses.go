package handlers

import (
	"net/http"

	"github.com/Myrtilli/transaction-indexing-svc/internal/service/models"
	"github.com/Myrtilli/transaction-indexing-svc/internal/service/requests"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GetAddresses(w http.ResponseWriter, r *http.Request) {
	logger := Log(r)
	username := Username(r)

	req, err := requests.Pagination(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	user, err := DB(r).User().GetByUsername(username)
	if err != nil {
		logger.WithError(err).Error("failed to get user")
		ape.RenderErr(w, problems.InternalError())
		return
	}
	if user == nil {
		ape.RenderErr(w, problems.NotFound())
		return
	}

	addresses, err := DB(r).Address().Select(user.ID, req.Pagination)
	if err != nil {
		logger.WithError(err).Error("failed to select addresses")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	ape.Render(w, models.AddressList(addresses))
}
