package app

import (
	"bytes"
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"

	"enroll/token"
)

type App struct {
	users       UserRepository
	encodingKey crypto.PrivateKey
}

// New creates a new app.
func New(users UserRepository, encodingKey crypto.PrivateKey) App {
	return App{users: users, encodingKey: encodingKey}
}

// Start the HTTP server.
func (app *App) Start() error {
	const serverAddr = "127.0.0.1:8080"

	r := mux.NewRouter()
	r.HandleFunc("/session", app.createSession).Methods(http.MethodPost)
	r.HandleFunc("/session/verification", app.validateCode).Methods(http.MethodPost)

	slog.Info(fmt.Sprintf("KommKorp enrollment application: http://%s\n", serverAddr))

	return http.ListenAndServe(serverAddr, r)
}

type SessionRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SessionResponse struct {
	VerifyRequestID string `json:"verify_request_id"`
}

// Handler for the session creation endpoint
func (app *App) createSession(w http.ResponseWriter, r *http.Request) {
	var req SessionRequest
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

	requestID, err := app.requestVerification(user)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(SessionResponse{VerifyRequestID: *requestID})
	if err != nil {
		slog.Error("Encoding failure: create session response", "error", err)
	}
}

type ValidateCodeRequest struct {
	Username  string `json:"username"`
	RequestID string `json:"request_id"`
	Code      string `json:"code"`
}

type ValidateCodeResponse struct {
	Token string `json:"token"`
}

// validateCode handles the verification of the code sent via SMS
func (app *App) validateCode(w http.ResponseWriter, r *http.Request) {
	var req ValidateCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("Failed to decode validation request", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.RequestID == "" || req.Code == "" {
		slog.Warn("Validation request missing fields", "username_present", req.Username != "", "request_id_present", req.RequestID != "", "code_present", req.Code != "")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	token, err := app.checkCode(req.Username, req.RequestID, req.Code)
	if err != nil {
		slog.Warn("Code validation failed", "error", err)
		http.Error(w, "invalid code", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(ValidateCodeResponse{Token: *token}); err != nil {
		slog.Error("Encoding failure: validate code response", "error", err)
	}
}

// request verification uses Vonage's Verify v2 API to request an authentication code via SMS
func (app *App) requestVerification(user User) (*string, error) {
	vonageToken, err := token.ProduceVonageToken(app.encodingKey)
	if err != nil {
		slog.Error("Failed to produce Vonage token", "error", err)
		return nil, fmt.Errorf("failed to produce vonage token: %w", err)
	}

	type verifyWorkflow struct {
		Channel string `json:"channel"`
		To      string `json:"to"`
	}

	type verifyRequest struct {
		Brand          string           `json:"brand"`
		ChannelTimeout uint             `json:"channel_timeout"`
		Workflow       []verifyWorkflow `json:"workflow"`
	}

	verifyReq := verifyRequest{
		Brand:          "KommKorp",
		ChannelTimeout: 300,
		Workflow: []verifyWorkflow{
			{Channel: "sms", To: user.Phone},
		},
	}

	verifyURL := "https://api.nexmo.com/v2/verify"
	body, _ := json.Marshal(verifyReq)
	reqBody, err := http.NewRequest(http.MethodPost, verifyURL, bytes.NewBuffer(body))
	if err != nil {
		slog.Error("Failed to create Verify HTTP request", "error", err)
		return nil, fmt.Errorf("failed to create verify http request: %w", err)
	}
	reqBody.Header.Set("Content-Type", "application/json")
	reqBody.Header.Set("Authorization", fmt.Sprintf("Bearer %s", vonageToken))

	client := &http.Client{}
	resp, err := client.Do(reqBody)
	if err != nil {
		slog.Error("Failed to call Vonage Verify API", "error", err)
		return nil, fmt.Errorf("failed to call vonage verify api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read Vonage Verify response body", "error", err)
		return nil, fmt.Errorf("failed to call vonage verify api: %w", err)
	}

	if resp.StatusCode != http.StatusAccepted {
		slog.Warn("Vonage Verify API returned non-OK status", "status", resp.StatusCode, "body", string(respBody))
		return nil, fmt.Errorf("failed to call vonage verify api: %w", err)
	}

	var verifyResponse struct {
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(respBody, &verifyResponse); err != nil {
		slog.Error("Failed to decode Vonage Verify response JSON", "error", err)
		return nil, fmt.Errorf("failed to decode vonage verify response: %w", err)
	}

	return &verifyResponse.RequestID, nil
}

// checkCode uses Vonage's Verify v2 API to check the provided code
func (app *App) checkCode(username, requestID, code string) (*string, error) {
	vonageToken, err := token.ProduceVonageToken(app.encodingKey)
	if err != nil {
		return nil, fmt.Errorf("failed to produce vonage token: %w", err)
	}

	checkURL := fmt.Sprintf("https://api.nexmo.com/v2/verify/%s", requestID)
	body, _ := json.Marshal(map[string]string{
		"code": code,
	})
	reqBody, err := http.NewRequest(http.MethodPost, checkURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create verify check http request: %w", err)
	}
	reqBody.Header.Set("Content-Type", "application/json")
	reqBody.Header.Set("Authorization", fmt.Sprintf("Bearer %s", vonageToken))

	client := &http.Client{}
	resp, err := client.Do(reqBody)
	if err != nil {
		slog.Error("Failed to call Vonage Verify check code API", "error", err)
		return nil, fmt.Errorf("failed to call vonage verify check api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read Vonage Verify response body", "error", err)
		return nil, fmt.Errorf("failed to read vonage verify check response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Warn("Vonage Verify API returned non-OK status", "status", resp.StatusCode, "body", string(respBody))
		return nil, fmt.Errorf("vonage verify check returned non-OK status: %w", err)
	}

	var checkResp struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(respBody, &checkResp); err != nil {
		slog.Error("Failed to decode Vonage Verify check response JSON", "error", err)
		return nil, fmt.Errorf("failed to decode vonage verify check response: %w", err)
	}

	if checkResp.Status != "completed" {
		slog.Error("Vonage Verify check failed", "status", checkResp.Status, "error", err)
		return nil, fmt.Errorf("vonage verify check failed with status: %s", checkResp.Status)
	}

	token, err := token.ProduceSessionToken(app.encodingKey, username)
	if err != nil {
		slog.Error("Failed to produce session token", "error", err)
		return nil, fmt.Errorf("failed to produce session token: %w", err)
	}

	return &token, nil
}
