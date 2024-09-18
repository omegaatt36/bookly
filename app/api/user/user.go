package user

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/service/user"
)

// RegisterUserRouters registers user-related routes on the provided router.
func (x *Controller) RegisterUserRouters(router *http.ServeMux) {
	router.HandleFunc("POST /users", x.createUserHandler)
	router.HandleFunc("GET /users", x.getAllUsersHandler)
	router.HandleFunc("GET /users/{id}", x.getUserByIDHandler)
	router.HandleFunc("PATCH /users/{id}", x.updateUserHandler)
	router.HandleFunc("DELETE /users/{id}", x.deactivateUserByIDHandler)
}

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

func (x *Controller) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Nickname string `json:"nickname"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := user.NewService(x.userRepo).CreateUser(domain.CreateUserRequest{
		Name:     req.Name,
		Nickname: req.Nickname,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func (x *Controller) getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := user.NewService(x.userRepo).GetAllUsers()
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

func (x *Controller) getUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "parameter 'id' is required", http.StatusBadRequest)
		return
	}

	u, err := user.NewService(x.userRepo).GetUserByID(id)
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

func (x *Controller) updateUserHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := user.NewService(x.userRepo).UpdateUser(domain.UpdateUserRequest{
		ID:       id,
		Name:     userName,
		Nickname: userNickname,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (x *Controller) deactivateUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "parameter 'id' is required", http.StatusBadRequest)
		return
	}

	err := user.NewService(x.userRepo).DeactivateUserByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
