package app

import (
	"fmt"
	"log/slog"
	"net/http"
)

type App struct {
}

// Create a new app.
func New() App {
	return App{}
}

// Start the HTTP server.
func (app *App) Start() error {
	const SERVER_ADDR = "127.0.0.1:8080"

	slog.Info(fmt.Sprintf("KommKorp enrollment application: http://%s\n", SERVER_ADDR))

	return http.ListenAndServe(SERVER_ADDR, nil)
}
