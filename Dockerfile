FROM golang:1.21 as build

WORKDIR /go/src/app

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
ENV CGO_ENABLED=0

COPY . .
RUN go build -ldflags="-s -w" -o /go/bin/app main.go

FROM gcr.io/distroless/base-debian11
WORKDIR /service/

COPY --from=build /go/bin/app .
COPY settings.yaml settings.yaml
CMD ["./app", "serve"]
