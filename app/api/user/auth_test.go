package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/omegaatt36/bookly/app/api/user"
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/database"
	"github.com/omegaatt36/bookly/persistence/repository"
	"github.com/omegaatt36/bookly/persistence/sqlc"
	"github.com/omegaatt36/bookly/service/auth"
)

type testAuthSuite struct {
	suite.Suite

	router *http.ServeMux

	repo     *repository.SQLCRepository
	finalize func()

	authenticator domain.Authenticator
	userID        int32
}

func (s *testAuthSuite) SetupTest() {
	s.finalize = database.TestingInitialize(database.PostgresOpt)
	db := database.GetDB()
	s.repo = repository.NewSQLCRepository(db)
	s.router = http.NewServeMux()
	s.authenticator = auth.NewJWTAuthorizator("salt", "secret")
	controller := user.NewController(s.repo,
		user.WithAuthenticator(domain.IdentityProviderPassword, s.authenticator),
	)

	s.router.HandleFunc("POST /auth/register", controller.RegisterUser())
	s.router.HandleFunc("POST /auth/login", controller.LoginUser())

	s.NoError(sqlc.MigrateForTest(context.Background(), db))
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
	email := "test@example.com"
	reqBody := []byte(fmt.Sprintf(`{"email": "%s", "password": "password"}`, email))

	s.T().Run("Register", func(_ *testing.T) {
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

		s.userID = user.ID
	})

	type loginResponse struct {
		Code int `json:"code"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}

	s.T().Run("Login", func(_ *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		s.router.ServeHTTP(w, req)
		s.Equal(http.StatusOK, w.Code)

		var resp loginResponse
		s.NoError(json.NewDecoder(w.Body).Decode(&resp))

		s.NotEmpty(resp.Data.Token)

		result, err := s.authenticator.ValidateToken(domain.ValidateTokenRequest{
			Token: resp.Data.Token,
		})

		s.NoError(err)
		s.True(result.Valid)
		user, _, err := s.repo.GetUserByIdentity(domain.IdentityProviderPassword, email)
		s.NoError(err)
		s.Equal(user.ID, result.UserID)
	})
}
