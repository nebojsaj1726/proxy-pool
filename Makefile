BINARY_NAME=proxy-pool

.PHONY: web

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

web:
	@if ! lsof -i:8080 > /dev/null; then \
		echo "API server not running on port 8080. Run 'make run' first."; \
	fi
	cd web && [ -d node_modules ] || npm install
	cd web && npm run dev
