default: test lint

test:
	go test ./...

lint:
	golangci-lint run

build:
	go build
