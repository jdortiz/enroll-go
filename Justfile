alias b := build
alias r := run
alias d := run_debug
alias t := test


# build project
build:
    go build -v -ldflags="-X main.version=$(git describe --always --long --dirty)" -o enroll cmd/enroll/main.go

# run server
run:
    go run -v -ldflags="-X main.version=$(git describe --always --long --dirty)" cmd/enroll/main.go

# run server (debug mode)
run_debug:
    GO_LOG=debug go run -v -ldflags="-X main.version=$(git describe --always --long --dirty)" cmd/enroll/main.go

# test all packages
test:
    go test ./...
