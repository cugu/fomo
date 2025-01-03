export CGO_ENABLED = 0
go = gotip

.PHONY: generate
generate:
	@echo "Generating..."
	cd db && go tool sqlc generate && cd ..

.PHONY: fmt
fmt:
	@echo "Formatting..."
	$(go) mod tidy
	$(go) fmt ./...
	$(go) tool gci write -s standard -s default -s "prefix(github.com/cugu/fomo)" .
	$(go) tool gofumpt -l -w .
	$(go) tool wsl -fix ./... || true

.PHONY: lint
lint:
	@echo "Linting..."
	$(go) tool golangci-lint run
	@echo "Done."

.PHONY: test
test:
	@echo "Testing..."
	$(go) test -v ./...
	@echo "Done."

.PHONY: run
run:
	@echo "Running..."
	$(go) run .
	@echo "Done."