package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/argon2"

	"github.com/omegaatt36/bookly/domain"
)

var _ domain.Authenticator = (*JWTAuthenticator)(nil)

// JWTAuthenticator represents a JWT authentication service, encrypts password by using argon2.
type JWTAuthenticator struct {
	ttl       time.Duration
	salt      string
	secretKey string

	getNow func() time.Time
}

// NewJWTAuthorizator creates a new JWT authentication service.
// Accept userRepo as the first parameter
func NewJWTAuthorizator(salt, secretKey string, opts ...Option) *JWTAuthenticator {
	authorizator := JWTAuthenticator{
		ttl:       time.Hour * 24,
		salt:      salt,
		secretKey: secretKey,
		getNow:    time.Now,
	}

	for _, opt := range opts {
		opt.apply(&authorizator)
	}

	return &authorizator
}

// Option represents an option for JWTAuthorizator.
type Option interface {
	apply(*JWTAuthenticator)
}

// WithTTL sets the time-to-live for tokens.
func WithTTL(ttl time.Duration) Option {
	return ttlOption{ttl: ttl}
}

type ttlOption struct {
	ttl time.Duration
}

func (o ttlOption) apply(a *JWTAuthenticator) {
	a.ttl = o.ttl
}

// HashPassword hashes a password using argon2.
func (authenticator *JWTAuthenticator) HashPassword(password string) (string, error) {
	hash := argon2.IDKey([]byte(password), []byte(authenticator.salt), 1, 64*1024, 4, 32)
	return fmt.Sprintf("%x", hash), nil
}

// GenerateToken generates a token for a user.
func (authenticator *JWTAuthenticator) GenerateToken(req domain.GenerateTokenRequest) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("failed to parse claims")
	}

	claims["sub"] = req.UserID
	claims["user_id"] = req.UserID
	claims["exp"] = authenticator.getNow().Add(authenticator.ttl).Unix()

	tokenString, err := token.SignedString([]byte(authenticator.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a token for a user.
func (authenticator *JWTAuthenticator) ValidateToken(req domain.ValidateTokenRequest) (bool, error) {
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(authenticator.secretKey), nil
	})

	if err != nil {
		return false, err
	}

	_, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, errors.New("invalid claims")
	}

	if !token.Valid {
		return false, errors.New("invalid token")
	}

	return true, nil
}

// VerifyCredential verifies the provided credential against the stored identity credential.
// It assumes the identity.Credential is the hex-encoded Argon2 hash and the salt is in authenticator.salt.

// VerifyCredential verifies if the provided credential matches the stored identity credential.
// It is used during authentication to validate user credentials.
func (authenticator *JWTAuthenticator) VerifyCredential(credential string, identity *domain.Identity) (bool, error) {
	if identity.Provider != domain.IdentityProviderPassword {
		return false, fmt.Errorf("unsupported identity provider for credential verification: %s", identity.Provider)
	}

	providedCredentialHash := argon2.IDKey([]byte(credential), []byte(authenticator.salt), 1, 64*1024, 4, 32)
	providedCredentialHashHex := fmt.Sprintf("%x", providedCredentialHash)

	if providedCredentialHashHex != identity.Credential {
		return false, nil
	}

	return true, nil
}
