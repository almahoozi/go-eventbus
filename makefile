.DEFAULT_GOAL := all

.PHONY: all init tools check lint test test_short report

all: test

init: tools
	@echo "Initializing..."
	@cp -r .githooks .git/hooks
	@chmod +x .git/hooks/*
# TODO: Do something else to ensure on the remote that the user has indeed initialized the project

tools:
	@echo "Installing tools..."
	@which tparse > /dev/null || { echo "Installing tparse..."; go install github.com/mfridman/tparse@latest; }

check: lint test_short report
	@git diff --exit-code -- TEST_REPORT.md || (echo "TEST_REPORT.md has unstaged changes. Please stage them and commit them." && exit 1)

lint:
	@echo "Linting..."
	@golangci-lint run

test: tools
	@echo "Running tests..."
	@go test -v ./... -json -cover -count=1 | tparse -follow -all

test_short:
	@echo "Running short tests..."
	@go test ./... -short -failfast > /dev/null || (echo "FAILED. Run 'make test' for more details." && exit 1)

report:
	@echo "Generating test report..."
	@go test -v ./... -json -cover | tparse -notests -format markdown > TEST_REPORT.md
