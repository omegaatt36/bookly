package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/service/auth"
)

type testAuthSuite struct {
	suite.Suite

	salt      string
	secretKey string
	// authorizator domain.Authenticator
}

func (s *testAuthSuite) SetupTest() {
	s.salt = "test-salt"
	s.secretKey = "test-secret-key"
}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(testAuthSuite))
}

func (s *testAuthSuite) TestHashPassword() {
	authenticator := auth.NewJWTAuthorizator(s.salt, s.secretKey)

	password := "test-password"
	hashedPassword, err := authenticator.HashPassword(password)

	s.NoError(err)
	s.NotEmpty(hashedPassword)
	s.NotEqual(password, hashedPassword)

	// Hash should be consistent
	hashedPassword2, err := authenticator.HashPassword(password)
	s.NoError(err)
	s.Equal(hashedPassword, hashedPassword2)
}

func (s *testAuthSuite) TestGenerateAndValidateToken() {
	authenticator := auth.NewJWTAuthorizator(s.salt, s.secretKey)
	userID := "test-user-id"
	token, err := authenticator.GenerateToken(domain.GenerateTokenRequest{UserID: userID})

	s.NoError(err)
	s.NotEmpty(token)

	// Validate the generated token
	valid, err := authenticator.ValidateToken(domain.ValidateTokenRequest{Token: token})
	s.NoError(err)
	s.True(valid)

	// Test with invalid token
	valid, err = authenticator.ValidateToken(domain.ValidateTokenRequest{Token: "invalid-token"})
	s.Error(err)
	s.False(valid)
}

func (s *testAuthSuite) TestTokenExpiration() {
	pastAuthenticator := auth.NewJWTAuthorizator(s.salt, s.secretKey, auth.WithTTL(time.Second), auth.WithGetNow(func() time.Time {
		return time.Now().Add(-time.Hour)
	}))

	userID := "test-user-id"

	token, err := pastAuthenticator.GenerateToken(domain.GenerateTokenRequest{UserID: userID})
	s.NoError(err)
	s.NotEmpty(token)

	normalAuthenticator := auth.NewJWTAuthorizator(s.salt, s.secretKey, auth.WithTTL(time.Second))

	valid, err := normalAuthenticator.ValidateToken(domain.ValidateTokenRequest{Token: token})
	s.Error(err)
	s.Equal("Token is expired", err.Error())
	s.False(valid)
}
