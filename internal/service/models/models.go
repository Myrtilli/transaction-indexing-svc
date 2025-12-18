package models

type SuccessResponse struct {
	Message string `json:"message"`
}

const (
	RegistrationSuccessMessage = "User registered successfully"
	LoginSuccessMessage        = "User logged in successfully"
	NewAddressSuccessMessage   = "Address added successfully"
)
