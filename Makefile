BINARY_NAME=proxy-pool


run:
	go run ./cmd/server


build:
	go build -o bin/$(BINARY_NAME) ./cmd/server


clean:
	rm -rf bin/


fmt:
	go fmt ./...	


lint:
	golangci-lint run || true


test:
	go test ./...