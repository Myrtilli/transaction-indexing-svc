package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Myrtilli/transaction-indexing-svc/internal/auth"
	"github.com/Myrtilli/transaction-indexing-svc/internal/data"
	passwordhash "github.com/Myrtilli/transaction-indexing-svc/internal/password_hash"
	"github.com/Myrtilli/transaction-indexing-svc/internal/service/models"
	"github.com/Myrtilli/transaction-indexing-svc/internal/service/requests"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func Register(w http.ResponseWriter, r *http.Request) {
	logger := Log(r)
	db := DB(r)
	key := JWTKey(r)
	time := JWTExpiration(r)

	var req requests.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("failed to decode request body", "error", err)
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	if err := req.Validate(); err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	hashedPassword, err := passwordhash.HashPassword(req.Password)
	if err != nil {
		logger.WithError(err).Error("failed to hash password")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	userInfo := data.User{
		Username:     req.Username,
		PasswordHash: hashedPassword,
	}

	_, err = db.User().Insert(userInfo)
	if err != nil {
		logger.WithError(err).Error("failed to insert user")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	user, err := db.User().GetByUsername(req.Username)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			ape.RenderErr(w, problems.Unauthorized())
			return
		}

		logger.WithError(err).Error("db error")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	token, err := auth.GenerateJWT(user.Username, []byte(key), time)
	if err != nil {
		Log(r).WithError(err).Error("failed to generate jwt")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	w.WriteHeader(http.StatusCreated)
	ape.Render(w, models.SuccessResponse{
		Message: models.RegistrationSuccessMessage,
		Token:   token,
	})
}
