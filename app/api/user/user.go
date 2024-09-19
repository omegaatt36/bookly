package user

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/omegaatt36/bookly/domain"
)

type jsonUser struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Name      string `json:"name"`
	Nickname  string `json:"nickname"`
	Disabled  bool   `json:"disabled"`
}

func (r *jsonUser) fromDomain(u *domain.User) {
	r.ID = u.ID
	r.CreatedAt = u.CreatedAt.Format(time.RFC3339)
	r.UpdatedAt = u.UpdatedAt.Format(time.RFC3339)
	r.Name = u.Name
	r.Nickname = u.Nickname
	r.Disabled = u.Disabled
}

// CreateUser handles the creation of a new user.
func (x *Controller) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Nickname string `json:"nickname"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := x.service.CreateUser(domain.CreateUserRequest{
		Name:     req.Name,
		Nickname: req.Nickname,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

// GetAllUsers retrieves all users from the system.
func (x *Controller) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := x.service.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonUsers := make([]jsonUser, len(users))
	for index, u := range users {
		jsonUsers[index].fromDomain(u)
	}

	bs, err := json.Marshal(jsonUsers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bs)
}

// GetUserByID retrieves a user by their ID.
func (x *Controller) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "parameter 'id' is required", http.StatusBadRequest)
		return
	}

	u, err := x.service.GetUserByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var jsonUser jsonUser
	jsonUser.fromDomain(u)

	bs, err := json.Marshal(jsonUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bs)
}

// UpdateUser handles updating a user's information.
func (x *Controller) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "parameter 'id' is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Name     *string `json:"name"`
		Nickname *string `json:"nickname"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var userName *string
	if req.Name != nil {
		userName = req.Name
	}
	var userNickname *string
	if req.Nickname != nil {
		userNickname = req.Nickname
	}

	if err := x.service.UpdateUser(domain.UpdateUserRequest{
		ID:       id,
		Name:     userName,
		Nickname: userNickname,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeactivateUserByID handles the deactivation of a user by their ID.
func (x *Controller) DeactivateUserByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "parameter 'id' is required", http.StatusBadRequest)
		return
	}

	err := x.service.DeactivateUserByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
