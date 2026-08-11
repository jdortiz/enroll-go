package app

import (
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"enroll/token"
	"github.com/gorilla/mux"
)

type App struct {
	users       UserRepository
	encodingKey crypto.PrivateKey
}

// Create a new app.
func New(users UserRepository, encodingKey crypto.PrivateKey) App {
	return App{users: users, encodingKey: encodingKey}
}

// Start the HTTP server.
func (app *App) Start() error {
	const SERVER_ADDR = "127.0.0.1:8080"

	r := mux.NewRouter()
	r.HandleFunc("/session", app.createSession).Methods(http.MethodPost)

	slog.Info(fmt.Sprintf("KommKorp enrollment application: http://%s\n", SERVER_ADDR))

	return http.ListenAndServe(SERVER_ADDR, r)
}

type sessionRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SessionResponse struct {
	JWT string `json:"jwt"`
}

func (app *App) createSession(w http.ResponseWriter, r *http.Request) {
	var req sessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("Failed to decode session request", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		slog.Warn("Session request missing credentials", "username_present", req.Username != "", "password_present", req.Password != "")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	user, err := app.users.FindByID(req.Username)
	if errors.Is(err, ErrUserNotFound) {
		slog.Info("Authentication failed: user not found", "username", req.Username)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		slog.Error("Database error during authentication", "username", req.Username, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if !user.VerifyPassword(req.Password) {
		slog.Info("Authentication failed: invalid password", "username", req.Username)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	jwt, err := token.ProduceSessionToken(app.encodingKey, user.Username)
	if err != nil {
		slog.Error("Failed to produce session token", "username", user.Username, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(SessionResponse{JWT: jwt})
	slog.Error("Encoding failure: create session response", "error", err)
}
