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

	"github.com/omegaatt36/bookly/app/api/engine"
	"github.com/omegaatt36/bookly/app/api/user"
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/database"
	"github.com/omegaatt36/bookly/persistence/repository"
	"github.com/omegaatt36/bookly/persistence/sqlc"
)

type testUserSuite struct {
	suite.Suite

	router *http.ServeMux

	repo     *repository.SQLCRepository
	finalize func()
	userID   int32
}

func (s *testUserSuite) SetupTest() {
	s.finalize = database.TestingInitialize(database.PostgresOpt)
	db := database.GetDB()
	s.repo = repository.NewSQLCRepository(db)
	s.router = http.NewServeMux()
	controller := user.NewController(s.repo)

	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := engine.WithUserID(r.Context(), s.userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	registerWithAuth := func(pattern string, handler http.Handler) {
		s.router.Handle(pattern, authMiddleware(handler))
	}

	s.router.HandleFunc("POST /users", controller.CreateUser()) // 不需要認證
	registerWithAuth("GET /users", http.HandlerFunc(controller.GetAllUsers()))
	registerWithAuth("GET /users/{id}", http.HandlerFunc(controller.GetUserByID()))
	registerWithAuth("PATCH /users/{id}", http.HandlerFunc(controller.UpdateUser()))
	registerWithAuth("DELETE /users/{id}", http.HandlerFunc(controller.DeactivateUserByID()))

	s.NoError(sqlc.MigrateForTest(context.Background(), db))
}

func (s *testUserSuite) TearDownTest() {
	s.finalize()
	s.router = nil
	s.repo = nil
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(testUserSuite))
}

func (s *testUserSuite) TestCreateUser() {
	type createUserResponse struct {
		Code int `json:"code"`
		Data any `json:"data"`
	}

	reqBody := []byte(`{"name": "Test User", "nickname": "test"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var resp createUserResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code)

	users, err := s.repo.GetAllUsers()
	s.NoError(err)
	s.Len(users, 1)

	user := users[0]
	s.Equal("Test User", user.Name)
	s.Equal("test", user.Nickname)
	s.False(user.Disabled)
}

func (s *testUserSuite) TestGetAllUsers() {
	type getAllUsersResponse struct {
		Code int `json:"code"`
		Data []struct {
			ID       int32  `json:"id"`
			Name     string `json:"name"`
			Nickname string `json:"nickname"`
			Disabled bool   `json:"disabled"`
		} `json:"data"`
	}

	// Create a test user
	userID, err := s.repo.CreateUser(domain.CreateUserRequest{
		Name:     "Test User",
		Nickname: "test",
	})
	s.NoError(err)

	s.userID = userID

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp getAllUsersResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code)
	s.Len(resp.Data, 1)
	s.Equal("Test User", resp.Data[0].Name)
	s.Equal("test", resp.Data[0].Nickname)
	s.False(resp.Data[0].Disabled)
}

func (s *testUserSuite) TestGetUserByID() {

	type getUserByIDResponse struct {
		Code int `json:"code"`
		Data struct {
			ID       int32  `json:"id"`
			Name     string `json:"name"`
			Nickname string `json:"nickname"`
			Disabled bool   `json:"disabled"`
		} `json:"data"`
	}

	// Create a test user
	userID, err := s.repo.CreateUser(domain.CreateUserRequest{Name: "Test User", Nickname: "test"})
	s.NoError(err)

	s.userID = userID

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%d", userID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp getUserByIDResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code)
	s.Equal(userID, resp.Data.ID)
	s.Equal("Test User", resp.Data.Name)
	s.Equal("test", resp.Data.Nickname)
	s.False(resp.Data.Disabled)
}

func (s *testUserSuite) TestUpdateUser() {
	type updateUserResponse struct {
		Code int `json:"code"`
	}
	// Create a test user
	userID, err := s.repo.CreateUser(domain.CreateUserRequest{Name: "Test User", Nickname: "test"})
	s.NoError(err)

	s.userID = userID

	reqBody := []byte(`{"name": "Updated User"}`)
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/users/%d", userID), bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp updateUserResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code)

	updatedUser, err := s.repo.GetUserByID(userID)
	s.NoError(err)
	s.Equal("Updated User", updatedUser.Name)
}

func (s *testUserSuite) TestDeactivateUserByID() {
	type deactivateUserResponse struct {
		Code int `json:"code"`
		Data any `json:"data"`
	}
	// Create a test user
	userID, err := s.repo.CreateUser(domain.CreateUserRequest{Name: "Test User", Nickname: "test"})
	s.NoError(err)

	s.userID = userID

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%d", userID), nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var resp deactivateUserResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code)

	deactivatedUser, err := s.repo.GetUserByID(userID)
	s.NoError(err)
	s.True(deactivatedUser.Disabled)
}
