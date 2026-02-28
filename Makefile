.PHONY: test test-race goimports lint

test:
	go test ./...

test-race:
	go test -race ./...

goimports:
	goimports -l -w .
	@echo "goimports done"

lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed, skip"
