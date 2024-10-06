package user_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/omegaatt36/bookly/app/api/user"
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/database"
	"github.com/omegaatt36/bookly/persistence/repository"
)

type testUserSuite struct {
	suite.Suite

	router *http.ServeMux

	repo     *repository.GORMRepository
	finalize func()
}

func (s *testUserSuite) SetupTest() {
	s.finalize = database.TestingInitialize(database.PostgresOpt)
	s.repo = repository.NewGORMRepository(database.GetDB())
	s.router = http.NewServeMux()
	controller := user.NewController(s.repo)
	s.router.HandleFunc("POST /users", controller.CreateUser())
	s.router.HandleFunc("GET /users", controller.GetAllUsers())
	s.router.HandleFunc("GET /users/{id}", controller.GetUserByID())
	s.router.HandleFunc("PATCH /users/{id}", controller.UpdateUser())
	s.router.HandleFunc("DELETE /users/{id}", controller.DeactivateUserByID())

	s.NoError(s.repo.AutoMigrate())
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
			ID       string `json:"id"`
			Name     string `json:"name"`
			Nickname string `json:"nickname"`
			Disabled bool   `json:"disabled"`
		} `json:"data"`
	}

	// Create a test user
	_, err := s.repo.CreateUser(domain.CreateUserRequest{
		Name:     "Test User",
		Nickname: "test",
	})
	s.NoError(err)

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
			ID       string `json:"id"`
			Name     string `json:"name"`
			Nickname string `json:"nickname"`
			Disabled bool   `json:"disabled"`
		} `json:"data"`
	}

	// Create a test user
	userID, err := s.repo.CreateUser(domain.CreateUserRequest{Name: "Test User", Nickname: "test"})
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/users/"+userID, nil)
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

	reqBody := []byte(`{"name": "Updated User"}`)
	req := httptest.NewRequest(http.MethodPatch, "/users/"+userID, bytes.NewBuffer(reqBody))
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

	req := httptest.NewRequest(http.MethodDelete, "/users/"+userID, nil)
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
