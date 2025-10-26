BINARY_NAME=proxy-pool


run:
	go run .


build:
	go build -o bin/$(BINARY_NAME) .


clean:
	rm -rf bin/


fmt:
	go fmt ./...	


lint:
	golangci-lint run || true


test:
	go test ./...