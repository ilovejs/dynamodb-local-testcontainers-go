#!/bin/sh

.PHONY: lint
lint:
	@echo "==> Run golangci-lint...";
	GOGC=10 golangci-lint run

.PHONY: test
test:
	@go test -timeout 300s -v 2>&1 ./... -coverpkg=./... -coverprofile=coverage.out -covermode count
	@go tool cover -html=coverage.out -o coverage.html

.PHONY: tidy
tidy:
	@go mod tidy -v

.PHONY: vet
vet:
	@go vet ./...
