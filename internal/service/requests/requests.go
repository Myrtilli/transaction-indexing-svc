package requests

import (
	"errors"
	"strings"
	"unicode"
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

func (r *NewAddressRequest) Validate() error {
	address := strings.TrimSpace(r.Address)
	if address == "" {
		return errors.New("address is required")
	}

	if len(address) < 26 {
		return errors.New("invalid address format")
	}
	return nil
}
