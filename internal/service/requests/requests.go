package requests

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/Myrtilli/transaction-indexing-svc/internal/data"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *RegisterRequest) Validate() error {
	username := strings.TrimSpace(r.Username)
	if username == "" {
		return errors.New("username is required")
	}

	if len(username) < 2 || len(username) > 16 {
		return errors.New("username must be between 2 and 16 characters")
	}

	for _, char := range username {
		if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
			return errors.New("username can only contain letters and numbers")
		}
	}

	return nil
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *LoginRequest) Validate() error {
	if strings.TrimSpace(r.Username) == "" {
		return errors.New("username is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

type NewAddressRequest struct {
	Address string `json:"address"`
}

func (r NewAddressRequest) Validate() error {
	if strings.TrimSpace(r.Address) == "" {
		return errors.New("address cannot be empty")
	}
	if len(r.Address) < 26 {
		return errors.New("address is too short")
	}
	return nil
}

type NewPaginationRequest struct {
	data.Pagination
}

func Pagination(r *http.Request) (NewPaginationRequest, error) {
	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	if limit <= 0 {
		limit = 15
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return NewPaginationRequest{
		Pagination: data.Pagination{
			Limit:  uint64(limit),
			Offset: uint64(offset),
		},
	}, nil
}

type NewTransactionRequest struct {
	Before *time.Time
	After  *time.Time
}

func TimeFilter(r *http.Request) (NewTransactionRequest, error) {
	q := r.URL.Query()
	var req NewTransactionRequest

	if s := q.Get("before"); s != "" {
		sec, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return req, errors.New("invalid 'before' timestamp")
		}
		t := time.Unix(sec, 0)
		req.Before = &t
	}

	if s := q.Get("after"); s != "" {
		sec, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return req, errors.New("invalid 'after' timestamp")
		}
		t := time.Unix(sec, 0)
		req.After = &t
	}

	return req, nil
}
