package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := Log(r)
		key := JWTKey(r)

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Warn("authentication failed: missing authorization header")
			ape.RenderErr(w, problems.Unauthorized())
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("authentication failed: invalid authorization format")
			ape.RenderErr(w, problems.Unauthorized())
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(key), nil
		})

		if err != nil || !token.Valid {
			logger.WithError(err).Error("authentication failed: token is invalid or expired", "error", err)
			ape.RenderErr(w, problems.Unauthorized())
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logger.Error("authentication failed: could not parse claims")
			ape.RenderErr(w, problems.Unauthorized())
			return
		}

		username, ok := claims["username"].(string)
		if !ok {
			logger.Error("authentication failed: username not found in token claims")
			ape.RenderErr(w, problems.Unauthorized())
			return
		}

		db := DB(r)
		user, err := db.User().GetByUsername(username)
		if err != nil || user == nil {
			Log(r).Error("user not found in database")
			ape.RenderErr(w, problems.Unauthorized())
			return
		}

		ctx := context.WithValue(r.Context(), userIDCtxKey, user.ID)

		ctx = context.WithValue(ctx, usernameCtxKey, username)

		authLogger := logger.WithField("auth_user", username)
		ctx = context.WithValue(ctx, logCtxKey, authLogger)

		authLogger.Debug("user authenticated successfully")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
