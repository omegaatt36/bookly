package user_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/omegaatt36/bookly/app/api/user"
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/database"
	"github.com/omegaatt36/bookly/persistence/repository"
	"github.com/omegaatt36/bookly/service/auth"
)

type testAuthSuite struct {
	suite.Suite

	router *http.ServeMux

	repo     *repository.GORMRepository
	finalize func()

	authenticator domain.Authenticator
}

func (s *testAuthSuite) SetupTest() {
	s.finalize = database.TestingInitialize(database.PostgresOpt)
	s.repo = repository.NewGORMRepository(database.GetDB())
	s.router = http.NewServeMux()
	s.authenticator = auth.NewJWTAuthorizator("salt", "secret")
	controller := user.NewController(s.repo,
		user.WithAuthenticator(domain.IdentityProviderPassword, s.authenticator),
	)

	s.router.HandleFunc("POST /auth/register", controller.RegisterUser)
	s.router.HandleFunc("POST /auth/login", controller.LoginUser)

	s.NoError(s.repo.AutoMigrate())
}

func (s *testAuthSuite) TearDownTest() {
	s.finalize()
	s.router = nil
	s.repo = nil
}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(testAuthSuite))
}

func (s *testAuthSuite) TestRegisterAndLogin() {
	reqBody := []byte(`{"email": "test@example.com", "password": "password123"}`)

	s.T().Run("Register", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		s.router.ServeHTTP(w, req)

		s.Equal(http.StatusCreated, w.Code)

		users, err := s.repo.GetAllUsers()
		s.NoError(err)
		s.Len(users, 1)

		user := users[0]
		s.Equal("test@example.com", user.Name)
		s.False(user.Disabled)
	})

	s.T().Run("Login", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		s.router.ServeHTTP(w, req)

		s.Equal(http.StatusOK, w.Code)

		token := w.Body.String()
		s.NotEmpty(token)

		valid, err := s.authenticator.ValidateToken(domain.ValidateTokenRequest{
			Token: token,
		})

		s.NoError(err)
		s.True(valid)
	})
}
