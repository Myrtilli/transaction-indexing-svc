package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Myrtilli/transaction-indexing-svc/internal/auth"
	passwordhash "github.com/Myrtilli/transaction-indexing-svc/internal/password_hash"
	"github.com/Myrtilli/transaction-indexing-svc/internal/service/requests"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func Login(w http.ResponseWriter, r *http.Request) {
	logger := Log(r)
	db := DB(r)

	var req requests.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("failed to decode request body", "error", err)
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	if err := req.Validate(); err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	user, err := db.User().Get(req.Username)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			ape.RenderErr(w, problems.Unauthorized())
			return
		}

		logger.WithError(err).Error("db error")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if !passwordhash.VerifyPassword(req.Password, user.PasswordHash) {
		logger.Error("invalid password")
		ape.RenderErr(w, problems.Unauthorized())
		return
	}

	token, err := auth.GenerateJWT(user.Username)
	if err != nil {
		Log(r).WithError(err).Error("failed to generate jwt")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	ape.Render(w, token)

}
