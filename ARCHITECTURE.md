# Brief Architecture Overview

Enroll is a server-side application written in Go.

Files and their function:
- `cmd/enroll/main.go` : Creates an instance of the application and launches it. It is also responsible of initializing
  the required infrastructure, like logging or reading environment files.
- `app/app.go` : Creates the app that defines the HTTP handlers and a router that binds them to the desired methods and
  URL. Additional handlers and functionality should be added here, but it also might be split into several files
  according to the handlers purpose, i.e. the (REST) resource.

