.PHONY: build clean run proto clean-proto clean-all deps build-prod build-full help

# Copy environment file from example (set to false for production)
COPY_ENV_FILE ?= true

# Setup PATH for protoc plugins
GOPATH_BIN := $(shell go env GOPATH)/bin
export PATH := $(PATH):$(GOPATH_BIN)

# Build the application (assumes proto files already exist)
build:
	@echo "Building atom engine..."
	go build -o build/atomd .
	@echo "Creating build directories..."
	mkdir -p build/config
	mkdir -p build/data/base
	@echo "Copying configuration files..."
	cp config/config.yaml build/config/config.yaml
ifeq ($(COPY_ENV_FILE),true)
	@echo "Copying environment example to .env..."
	cp config/env.example build/config/.env
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
	@echo "Protobuf generation completed"

# Show help
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
