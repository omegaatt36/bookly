package domain

// GenerateTokenRequest defines the request to generate a token
type GenerateTokenRequest struct {
	UserID string
}

// ValidateTokenRequest defines the request to validate a token
type ValidateTokenRequest struct {
	Token string
}

// TokenValidationResponse defines the response for token validation
type TokenValidationResponse struct {
	Valid  bool
	UserID string
}

// Authenticator represents an authentication service
type Authenticator interface {
	HashPassword(password string) (string, error)
	GenerateToken(GenerateTokenRequest) (string, error)
	ValidateToken(ValidateTokenRequest) (TokenValidationResponse, error)
	VerifyCredential(credential string, identity *Identity) (bool, error)
}
