golangci_version = v1.60.2
gofumpt_version = v0.6.0

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## audit: run all quality control checks
.PHONY: audit
audit: audit/tidy audit/format audit/vet audit/vulnerabilities audit/lint

## audit/tidy: check if go.mod is tidy
.PHONY: audit/tidy
audit/tidy:
	@echo 'Checking if go.mod is tidy:'
	@echo 
	go mod tidy -diff
	go mod verify
	@echo

## audit/format: check if code is formatted
.PHONY: audit/format
audit/format:
	@echo 'Checking if code is formated:'
	@echo 
	test -z "$(shell go run mvdan.cc/gofumpt@${gofumpt_version} -l -extra .)" 
	@echo

## audit/vet: check go vet
.PHONY: audit/vet
audit/vet:
	@echo 'Checking Go vet:'
	@echo 
	go vet ./...
	@echo

## audit/vulnerabilities: scan for vulnerabilities
.PHONY: audit/vulnerabilities
audit/vulnerabilities:
	@echo 'Scanning for vulnerabilities:'
	@echo 
	go run golang.org/x/vuln/cmd/govulncheck@latest -show verbose ./...
	@echo


## audit/lint: run linter
.PHONY: audit/lint
audit/lint:
	@echo 'Linting:'
	@echo
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@${golangci_version} run

## test: run all tests
.PHONY: test
test:
	go test -timeout 10s -race -v ./...

## test/cover: run tests with coverage -timeout 10s
.PHONY: test/cover
test/cover:
	go test -timeout 10s -race -v -coverprofile=coverage.txt ./...

## tidy: tidy modfiles, format code and run linter
.PHONY: tidy
tidy:
	go mod tidy
	go run mvdan.cc/gofumpt@${gofumpt_version} -w -extra .
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@${golangci_version} run --fix

