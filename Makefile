default: license test lint

test:
	go test ./...

lint:
	golangci-lint run

build:
	go build

license:
	./scripts/check_license.sh
