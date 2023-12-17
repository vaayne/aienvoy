
static:
	golangci-lint run --fix ./... ./examples/... ./cmd/...
	pre-commit run --all-files

illm:
	go build -ldflags="-s -w" -o illm cmd/illm/main.go && mv illm `go env GOPATH`/bin/illm

run: build
	./app serve
	rm app

build:
	go build -ldflags="-s -w" -o ./app main.go

# create migration file
migrate_create:
	# usage make migrate_create name=table_name
	go run main.go migrate create $(name)

# run migration up or down
# for example: make migrate action=up, make migrate action=down
migrate:
	go run main.go migrate $(action)
