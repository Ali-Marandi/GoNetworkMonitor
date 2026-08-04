APP_NAME    := gonetmon
VERSION     := 2.0.0
BUILD_DIR   := dist
LDFLAGS     := -ldflags "-X main.Version=$(VERSION) -s -w"
CGO_FLAGS   := CGO_ENABLED=1

.PHONY: all build clean test linux windows darwin release

all: build

build:
	@echo "Building $(APP_NAME) v$(VERSION) for current platform..."
	@mkdir -p $(BUILD_DIR)
	$(CGO_FLAGS) go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/gonetmon/
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

linux:
	@echo "Building for Linux (amd64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(CGO_FLAGS) go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./cmd/gonetmon/
	@chmod +x $(BUILD_DIR)/$(APP_NAME)-linux-amd64

linux-arm64:
	@echo "Building for Linux (arm64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 $(CGO_FLAGS) CC=aarch64-linux-gnu-gcc go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 ./cmd/gonetmon/

test:
	@echo "Running tests..."
	go test ./... -v -race -cover

vet:
	go vet ./...

fmt:
	gofmt -s -w .

clean:
	@rm -rf $(BUILD_DIR) gonetmon
	@echo "Cleaned build artifacts"

release: clean linux
	@echo "Creating release archives..."
	@mkdir -p $(BUILD_DIR)/release
	@cp README.md LICENSE $(BUILD_DIR)/
	@tar -czf $(BUILD_DIR)/release/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz \
		-C $(BUILD_DIR) $(APP_NAME)-linux-amd64 README.md LICENSE
	@echo "Release archives created in $(BUILD_DIR)/release/"

run:
	@echo "Starting GoNetworkMonitor..."
	sudo ./gonetmon --port 8080

docker-build:
	docker build -t gonetworkmonitor:$(VERSION) .
	docker tag gonetworkmonitor:$(VERSION) gonetworkmonitor:latest

help:
	@echo "GoNetworkMonitor v$(VERSION) — Build Targets"
	@echo ""
	@echo "  make build       Build for current platform"
	@echo "  make linux       Cross-compile for Linux amd64"
	@echo "  make test        Run all tests"
	@echo "  make clean       Remove build artifacts"
	@echo "  make release     Build and package for release"
	@echo "  make run         Run the application (requires sudo)"
