IMAGE_NAME=vaayne/aienvoy
IMAGE_TAG=latest
IMAGE_CACHE_TAG=buildcache

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

buildd:
	docker buildx build --cache-from=type=registry,ref=${IMAGE_NAME}:${IMAGE_CACHE_TAG} --cache-to=type=registry,ref=${IMAGE_NAME}:${IMAGE_CACHE_TAG},mode=max -t ${IMAGE_NAME}:${IMAGE_TAG} --load .

buildd-push:
	docker buildx build --platform linux/arm64,linux/amd64 --cache-from=type=registry,ref=${IMAGE_NAME}:${IMAGE_CACHE_TAG} --cache-to=type=registry,ref=${IMAGE_NAME}:${IMAGE_CACHE_TAG},mode=max -t ${IMAGE_NAME}:${IMAGE_TAG} --push .

# create migration file
migrate_create:
	# usage make migrate_create name=table_name
	go run main.go migrate create $(name)

# run migration up or down
# for example: make migrate action=up, make migrate action=down
migrate:
	go run main.go migrate $(action)
