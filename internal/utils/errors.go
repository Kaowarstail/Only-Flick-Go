package utils

import "errors"

// Erreurs communes pour l'authentification
var (
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenInvalid       = errors.New("token is invalid")
	ErrTokenNotFound      = errors.New("token not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrAccountDisabled    = errors.New("account is disabled")
	ErrAccountBanned      = errors.New("account is banned")
)
