# Go parameters
GOCMD?=go
GO_VERSION=$(shell go list -m -f "{{.GoVersion}}")
PACKAGE_BASE=github.com/buildpacks/libcnb/v2

all: test

install-goimports:
	@echo "> Installing goimports..."
	cd tools && $(GOCMD) install golang.org/x/tools/cmd/goimports

format: install-goimports
	@echo "> Formating code..."
	@goimports -l -w -local ${PACKAGE_BASE} .

install-golangci-lint:
	@echo "> Installing golangci-lint..."
	cd tools && $(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint

lint: install-golangci-lint
	@echo "> Linting code..."
	@golangci-lint run -c golangci.yaml

test: format lint
	$(GOCMD) test -parallel=1 -count=1 -v ./...
