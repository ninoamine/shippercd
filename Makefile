.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

.PHONY: lint
lint: fmt vet
	@echo "Linting code..."
	golangci-lint run ./...


.PHONY: test
test: lint
	@echo "Running tests..."
	go test -v ./internal/controllers/...