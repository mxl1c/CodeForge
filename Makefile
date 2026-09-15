.PHONY: build test tidy

build:
	go build -o bin/codeforge ./cmd/codeforge

test:
	go test ./...
	go test -C samples/go-saas-admin ./...

tidy:
	go mod tidy
