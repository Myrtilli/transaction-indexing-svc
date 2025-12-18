package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Myrtilli/transaction-indexing-svc/internal/data"
	"github.com/Myrtilli/transaction-indexing-svc/internal/service/models"
	"github.com/Myrtilli/transaction-indexing-svc/internal/service/requests"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func NewAddress(w http.ResponseWriter, r *http.Request) {
	var req requests.NewAddressRequest

	logger := Log(r)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("failed to decode request body")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	if err := req.Validate(); err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	username := Username(r)
	if username == "" {
		logger.Error("username not found in context")
		ape.RenderErr(w, problems.Unauthorized())
		return
	}

	user, err := DB(r).User().GetByUsername(username)
	if err != nil {
		logger.WithError(err).Error("failed to get user by username")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if user == nil {
		logger.Warn("user not found")
		ape.RenderErr(w, problems.NotFound())
		return
	}

	err = DB(r).Address().Insert(data.Address{
		UserID:  user.ID,
		Address: req.Address,
	})

	if err != nil {
		logger.WithError(err).Error("failed to insert new address")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	w.WriteHeader(http.StatusCreated)
	ape.Render(w, models.SuccessResponse{Message: models.NewAddressSuccessMessage})

}
