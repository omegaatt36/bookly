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
	s.router.HandleFunc("POST /users", controller.CreateUser)
	s.router.HandleFunc("GET /users", controller.GetAllUsers)
	s.router.HandleFunc("GET /users/{id}", controller.GetUserByID)
	s.router.HandleFunc("PATCH /users/{id}", controller.UpdateUser)
	s.router.HandleFunc("DELETE /users/{id}", controller.DeactivateUserByID)

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
	reqBody := []byte(`{"name": "Test User", "nickname": "test"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	users, err := s.repo.GetAllUsers()
	s.NoError(err)
	s.Len(users, 1)

	user := users[0]
	s.Equal("Test User", user.Name)
	s.Equal("test", user.Nickname)
	s.False(user.Disabled)
}

func (s *testUserSuite) TestGetAllUsers() {
	// Create a test user
	s.repo.CreateUser(domain.CreateUserRequest{
		Name:     "Test User",
		Nickname: "test",
	})

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var users []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Nickname string `json:"nickname"`
		Disabled bool   `json:"disabled"`
	}

	s.NoError(json.Unmarshal(w.Body.Bytes(), &users))
	s.Len(users, 1)
	s.Equal("Test User", users[0].Name)
	s.Equal("test", users[0].Nickname)
	s.False(users[0].Disabled)
}

func (s *testUserSuite) TestGetUserByID() {
	// Create a test user
	s.repo.CreateUser(domain.CreateUserRequest{Name: "Test User", Nickname: "test"})
	users, _ := s.repo.GetAllUsers()
	userID := users[0].ID

	req := httptest.NewRequest(http.MethodGet, "/users/"+userID, nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var user struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Nickname string `json:"nickname"`
		Disabled bool   `json:"disabled"`
	}

	s.NoError(json.Unmarshal(w.Body.Bytes(), &user))
	s.Equal(userID, user.ID)
	s.Equal("Test User", user.Name)
	s.Equal("test", user.Nickname)
	s.False(user.Disabled)
}

func (s *testUserSuite) TestUpdateUser() {
	// Create a test user
	s.repo.CreateUser(domain.CreateUserRequest{Name: "Test User", Nickname: "test"})
	users, _ := s.repo.GetAllUsers()
	userID := users[0].ID

	reqBody := []byte(`{"name": "Updated User"}`)
	req := httptest.NewRequest(http.MethodPatch, "/users/"+userID, bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	updatedUser, err := s.repo.GetUserByID(userID)
	s.NoError(err)
	s.Equal("Updated User", updatedUser.Name)
}

func (s *testUserSuite) TestDeactivateUserByID() {
	// Create a test user
	s.repo.CreateUser(domain.CreateUserRequest{Name: "Test User", Nickname: "test"})
	users, _ := s.repo.GetAllUsers()
	userID := users[0].ID

	req := httptest.NewRequest(http.MethodDelete, "/users/"+userID, nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	deactivatedUser, err := s.repo.GetUserByID(userID)
	s.NoError(err)
	s.True(deactivatedUser.Disabled)
}
