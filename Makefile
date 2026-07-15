.PHONY: run build test lint clean install-lint check dev help

BINARY_NAME=orchestrator
BINARY_DIR=bin

LINT_CMD := $(shell command -v golangci-lint 2>/dev/null)

install-lint:
	@echo "🔧 Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ installed. Add $(go env GOPATH)/bin to your PATH if needed."

build:
	@echo "🔨 Building..."
	go build -o $(BINARY_DIR)/$(BINARY_NAME) cmd/orchestrator/main.go

run: build
	@echo "🚀 Running..."
	$(BINARY_DIR)/$(BINARY_NAME)

dev:
	@echo "🚀 Running in dev mode..."
	go run cmd/orchestrator/main.go

test:
	@echo "🧪 Running tests..."
	go test ./... -v

lint:
	@if [ -z "$(LINT_CMD)" ]; then \
		echo "⚠️ golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		echo "✅ installed. Run 'make lint' again."; \
		exit 1; \
	fi
	@echo "🔍 Running linter..."
	golangci-lint run ./...

clean:
	@echo "🧹 Cleaning..."
	rm -rf $(BINARY_DIR)
	go clean

check: lint build
	@echo "✅ All checks passed"

help:
	@echo "Available targets:"
	@echo "  run          - Build and run orchestrator"
	@echo "  dev          - Run orchestrator via go run"
	@echo "  build        - Build binary"
	@echo "  test         - Run all tests"
	@echo "  lint         - Run linter (auto-install)"
	@echo "  clean        - Clean build artifacts"
	@echo "  install-lint - Install golangci-lint"
	@echo "  check        - Full check (lint + build)"

# Демонстрация: запуск оркестратора (фон) и эмулятора (на переднем плане)
demo:
	@echo "🚀 Starting orchestrator in background..."
	@go run cmd/orchestrator/main.go > orchestrator.log 2>&1 &
	@sleep 2
	@echo "🔧 Starting emulator (logs will appear below)..."
	@echo "📡 Orchestrator logs are in orchestrator.log"
	@echo "Press Ctrl+C to stop both."
	@trap 'kill %1; exit' INT; \
	go run emulator/cmd/emulator/main.go
	@echo "✅ Demo finished"

# Остановка демонстрации (если нужно вручную)
stop-demo:
	@echo "Stopping orchestrator..."
	-pkill -f "cmd/orchestrator/main.go"
	@echo "Stopped."