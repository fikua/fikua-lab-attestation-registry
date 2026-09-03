.PHONY: run build test fmt vet

run:
	go run ./cmd/registry

build:
	go build -trimpath -ldflags="-s -w" -o bin/registry ./cmd/registry

test:
	go test ./...

fmt:
	gofmt -l .

vet:
	go vet ./...
