# Brief Architecture Overview

Enroll is a server-side application written in Go.

Files and their function:
- `cmd/enroll/main.go` : Entry point; initializes infrastructure (logging, env) and starts the app.
- `app/app.go` : Defines the HTTP router and request handlers for the enrollment flow.
- `app/user.go` : Contains core domain models (`User`) and the `UserRepository` interface.
- `app/user_inmemory_repo.go` : A thread-safe, in-memory implementation of `UserRepository`.
- `token/token.go` : Handles JWT production, signature algorithm selection, and private key loading.

