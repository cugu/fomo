export CGO_ENABLED = 0

.PHONY: generate
generate:
	@echo "Generating..."
	cd db && go tool sqlc generate && cd ..

.PHONY: fmt
fmt:
	@echo "Formatting..."
	go mod tidy
	go fmt ./...
	go tool golangci-lint run --fix

.PHONY: lint
lint:
	@echo "Linting..."
	go tool golangci-lint run
	@echo "Done."

.PHONY: test
test:
	@echo "Testing..."
	go test -v ./...
	@echo "Done."

.PHONY: run
run:
	@echo "Running..."
	go run . -password=foobar
	@echo "Done."