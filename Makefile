.PHONY: build test lint run clean

build:
	go build -o bin/http-server ./cmd/...

test:
	go test -race -v ./...

lint:
	go vet ./...
	golangci-lint run

run: build
	./bin/http-server

clean:
	rm -rf bin/