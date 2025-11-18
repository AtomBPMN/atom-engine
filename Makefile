.PHONY: build clean run proto clean-proto clean-all deps build-prod build-full help lint lint-install

# Copy environment file from example (set to false for production)
COPY_ENV_FILE ?= true

# Setup PATH for protoc plugins
GOPATH_BIN := $(shell go env GOPATH)/bin
export PATH := $(PATH):$(GOPATH_BIN)

# Build variables
BASE_VERSION := $(shell cat version.txt 2>/dev/null || echo "dev")
BUILD_NUMBER_FILE := build_number.txt
CURRENT_BUILD := $(shell cat $(BUILD_NUMBER_FILE) 2>/dev/null || echo "0")
NEW_BUILD := $(shell expr $(CURRENT_BUILD) + 1)
VERSION := $(BASE_VERSION).$(NEW_BUILD)
GIT_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Build the application (assumes proto files already exist)
build:
	@echo "Building atom engine..."
	@echo "Base version: $(BASE_VERSION)"
	@echo "Current build: $(CURRENT_BUILD)"
	@echo "New build: $(NEW_BUILD)"
	@echo "Full version: $(VERSION)"
	@echo "Git commit: $(GIT_COMMIT)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Incrementing build number..."
	@echo $(NEW_BUILD) > $(BUILD_NUMBER_FILE)
	go build -ldflags "-X atom-engine/src/version.Version=$(VERSION) -X atom-engine/src/version.GitCommit=$(GIT_COMMIT) -X atom-engine/src/version.BuildTime=$(BUILD_TIME)" -o build/atomd .
	@echo "Creating build directories..."
	mkdir -p build/config
	mkdir -p build/data/base
	@echo "Copying configuration files..."
	@if [ -f config/config.yaml ]; then \
		echo "Using existing config/config.yaml..."; \
		cp config/config.yaml build/config/config.yaml; \
	else \
		echo "Using config/config.yaml.example..."; \
		cp config/config.yaml.example build/config/config.yaml; \
	fi
ifeq ($(COPY_ENV_FILE),true)
	@if [ -f config/.env ]; then \
		echo "Using existing config/.env..."; \
		cp config/.env build/config/.env; \
	else \
		echo "Using config/env.example..."; \
		cp config/env.example build/config/.env; \
	fi
else
	@echo "Skipping .env file copy (production mode)"
endif
	@echo "Build completed successfully"

# Build with proto generation (full build from scratch)
build-full: proto build
	@echo "Full build with proto generation completed"

# Clean build artifacts
clean:
	rm -rf build/

# Clean generated protobuf files
clean-proto:
	@echo "Cleaning generated protobuf files..."
	rm -rf proto/*/storagepb
	rm -rf proto/*/processpb
	rm -rf proto/*/parserpb
	rm -rf proto/*/timewheelpb
	rm -rf proto/*/jobspb
	rm -rf proto/*/messagespb
	rm -rf proto/*/expressionpb
	rm -rf proto/*/incidentspb
	@echo "Proto cleanup completed"

# Full clean (build + proto)
clean-all: clean clean-proto

# Run the application
run: build
	./build/atomd

# Install dependencies
deps:
	go mod tidy
	go mod download

# Build for production (without copying .env file)
build-prod:
	$(MAKE) build COPY_ENV_FILE=false

# Generate protobuf files
proto:
	@echo "Generating protobuf files..."
	@echo "Creating output directories..."
	mkdir -p proto/storage/storagepb
	mkdir -p proto/process/processpb  
	mkdir -p proto/parser/parserpb
	mkdir -p proto/timewheel/timewheelpb
	mkdir -p proto/jobs/jobspb
	mkdir -p proto/messages/messagespb
	mkdir -p proto/expression/expressionpb
	mkdir -p proto/incidents/incidentspb
	@echo "Generating storage proto..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/storage/storage.proto
	mv proto/storage/*.pb.go proto/storage/storagepb/ 2>/dev/null || true
	@echo "Generating process proto..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/process/process.proto
	mv proto/process/*.pb.go proto/process/processpb/ 2>/dev/null || true
	@echo "Generating parser proto..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/parser/parser.proto
	mv proto/parser/*.pb.go proto/parser/parserpb/ 2>/dev/null || true
	@echo "Generating timewheel proto..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/timewheel/timewheel.proto
	mv proto/timewheel/*.pb.go proto/timewheel/timewheelpb/ 2>/dev/null || true
	@echo "Generating jobs proto..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/jobs/jobs.proto
	mv proto/jobs/*.pb.go proto/jobs/jobspb/ 2>/dev/null || true
	@echo "Generating messages proto..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/messages/messages.proto
	mv proto/messages/*.pb.go proto/messages/messagespb/ 2>/dev/null || true
	@echo "Generating expression proto..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/expression/expression.proto
	mv proto/expression/*.pb.go proto/expression/expressionpb/ 2>/dev/null || true
	@echo "Generating incidents proto..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/incidents/incidents.proto
	mv proto/incidents/*.pb.go proto/incidents/incidentspb/ 2>/dev/null || true
	@echo "Protobuf generation completed"

# Run golangci-lint code analysis
lint:
	@echo "Checking for golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Error: golangci-lint is not installed."; \
		echo "Run 'make lint-install' for installation instructions."; \
		exit 1; \
	fi
	@if [ ! -f .golangci.yml ] && [ ! -f .golangci.yaml ]; then \
		echo "Creating .golangci.yml from example..."; \
		cp .golangci.yml.example .golangci.yml; \
		echo "Local configuration created. You can customize .golangci.yml for your needs."; \
	fi
	@echo "Running golangci-lint..."
	@golangci-lint run --max-issues-per-linter=0 --max-same-issues=0

# Show installation instructions for golangci-lint
lint-install:
	@echo "golangci-lint Installation Instructions"
	@echo "======================================="
	@echo ""
	@echo "Install using one of the following methods:"
	@echo ""
	@echo "1. Using curl (Linux/macOS):"
	@echo "   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin latest"
	@echo ""
	@echo "2. Using go install:"
	@echo "   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
	@echo ""
	@echo "3. Using Homebrew (macOS):"
	@echo "   brew install golangci-lint"
	@echo ""
	@echo "4. Using apt (Debian/Ubuntu):"
	@echo "   sudo apt-get install golangci-lint"
	@echo ""
	@echo "After installation, run 'make lint' to check your code."
	@echo ""
	@echo "Note: Local .golangci.yml configuration is ignored by git."
	@echo "      Copy .golangci.yml.example to .golangci.yml to customize."

# Show help
# Reset build number to 0
reset-build:
	@echo "Resetting build number to 0..."
	@echo "0" > $(BUILD_NUMBER_FILE)
	@echo "Build number reset to 0"

# Set base version (usage: make set-version VERSION=1.2)
set-version:
	@echo "Setting base version to $(VERSION)..."
	@echo "$(VERSION)" > version.txt
	@echo "Base version set to $(VERSION)"

# Show current version info
version-info:
	@echo "Base Version: $(BASE_VERSION)"
	@echo "Current Build: $(CURRENT_BUILD)"
	@echo "Next Build: $(NEW_BUILD)"
	@echo "Full Version: $(VERSION)"

help:
	@echo "Atom Engine Build System"
	@echo "======================="
	@echo ""
	@echo "Main commands:"
	@echo "  make build       - Build application (assumes proto files exist)"
	@echo "  make build-full  - Full build with proto generation" 
	@echo "  make run         - Build and run application"
	@echo ""
	@echo "Development commands:"
	@echo "  make proto       - Generate all protobuf files"
	@echo "  make deps        - Install/update dependencies"
	@echo "  make lint        - Run golangci-lint code analysis"
	@echo "  make lint-install - Show golangci-lint installation instructions"
	@echo ""
	@echo "Version commands:"
	@echo "  make version-info - Show current version information"
	@echo "  make set-version  - Set base version (usage: make set-version VERSION=1.2)"
	@echo "  make reset-build  - Reset build number to 0"
	@echo ""
	@echo "Cleanup commands:"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make clean-proto - Remove generated proto files"
	@echo "  make clean-all   - Remove build + proto files"
	@echo ""
	@echo "Production:"
	@echo "  make build-prod  - Build without .env file copying"
	@echo ""
	@echo "Notes:"
	@echo "  - PATH is automatically configured for protoc plugins"
	@echo "  - Go PATH: $(GOPATH_BIN)"
