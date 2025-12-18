package models

import "github.com/Myrtilli/transaction-indexing-svc/internal/data"

type SuccessResponse struct {
	Message string `json:"message"`
}

const (
	RegistrationSuccessMessage = "User registered successfully"
	LoginSuccessMessage        = "User logged in successfully"
	NewAddressSuccessMessage   = "Address added successfully"
)

type AddressModel struct {
	ID      int64  `json:"id"`
	Address string `json:"address"`
}

func AddressList(src []data.Address) []AddressModel {
	res := make([]AddressModel, len(src))
	for i, v := range src {
		res[i] = AddressModel{
			ID:      v.ID,
			Address: v.Address,
		}
	}
	return res
}
