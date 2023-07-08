GO_BUILD_CMD= CGO_ENABLED=0 go build

BINARY_NAME=dance
BUILD_DIR=build

.PHONY: all
all: clean lint build-all package-all

.PHONY: lint
lint:
	@echo "Linting code..."
	@go vet ./...

.PHONY: pre-build
pre-build:
	@mkdir -p $(BUILD_DIR)

.PHONY: build-linux
build-linux: pre-build
	@echo "Building Linux binary..."
	GOOS=linux GOARCH=amd64 $(GO_BUILD_CMD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64

.PHONY: build-osx
build-osx: pre-build
	@echo "Building OSX binary..."
	GOOS=darwin GOARCH=amd64 $(GO_BUILD_CMD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64

.PHONY: build build-all
build-all: build-linux build-osx

.PHONY: package-linux
package-linux:
	@echo "Packaging Linux binary..."
	tar -C $(BUILD_DIR) -zcf $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64

.PHONY: package-osx
package-osx:
	@echo "Packaging OSX binary..."
	tar -C $(BUILD_DIR) -zcf $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64

.PHONY: package-all
package-all: package-linux package-osx

.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -Rf $(BUILD_DIR)
