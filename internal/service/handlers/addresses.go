package handlers

import (
	"net/http"

	"github.com/Myrtilli/transaction-indexing-svc/internal/service/models"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GetAddresses(w http.ResponseWriter, r *http.Request) {
	logger := Log(r)
	username := Username(r)

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

	addresses, err := DB(r).Address().Select(user.ID)
	if err != nil {
		logger.WithError(err).Error("failed to select addresses")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	ape.Render(w, models.AddressList(addresses))
}
