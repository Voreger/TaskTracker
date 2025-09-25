package handlers

import (
	"GoProjects/TaskTracker/internal/auth"
	"GoProjects/TaskTracker/internal/models"
	"GoProjects/TaskTracker/internal/store"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type AuthHandler struct {
	Store *store.UserStore
}

func RegisterAuthRoutes(r chi.Router, s *store.UserStore) {
	h := &AuthHandler{Store: s}

	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)
}

// Register godoc
// @Summary      Register new user
// @Description  Creates a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body      models.User  true  "User info"
// @Success      201   {object}  models.User
// @Failure      400   {string}  string "invalid input"
// @Failure      500   {string}  string "internal error"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hashed, err := auth.HashPassword(u.Password)
	if err != nil {
		http.Error(w, "failed to hashed password", http.StatusInternalServerError)
		return
	}
	u.Password = hashed

	if err := h.Store.Create(context.Background(), &u); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(u)
}

// Login godoc
// @Summary      Login
// @Description  Authenticate user and return JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      map[string]string  true  "Email and password"
// @Success      200  {object}  map[string]string "token"
// @Failure      400  {string}  string "invalid input"
// @Failure      401  {string}  string "unauthorized"
// @Failure      500  {string}  string "internal error"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := h.Store.GetByEmail(context.Background(), req.Email)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if !auth.CheckPassword(user.Password, req.Password) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"token": token})
}
