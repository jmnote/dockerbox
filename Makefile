GOLANGCI_LINT_VER := v1.59.1
GO_LICENSES_VER := v1.6.0

tidy:
	go mod tidy
.PHONY: tidy

test:
	go test -v ./... -race -failfast
.PHONY: test

cover:
	go test -failfast -coverprofile=/tmp/profile.cov ./...
	go tool cover -func /tmp/profile.cov
.PHONY: cover

lint:
	go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VER) || true
	$(shell go env GOPATH)/bin/golangci-lint run
.PHONY: lint

licenses:
	go install -v github.com/google/go-licenses@$(GO_LICENSES_VER) || true
	$(shell go env GOPATH)/bin/go-licenses check ./box/
.PHONY: licenses

checks: tidy test lint licenses
.PHONY: checks