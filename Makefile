
static:
	golangci-lint run --fix --enable gofumpt
	pre-commit run --all-files

run:
	./app

build:
	go build -ldflags="-s -w" -o ./app main.go
