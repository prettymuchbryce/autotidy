.PHONY: build build-all install install-service uninstall uninstall-service clean test nix-update-deps

# Build variables
BINARY_NAME := autotidy
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_FLAGS := -ldflags "-X main.version=$(VERSION)"

# Detect OS
UNAME_S := $(shell uname -s)

# Default target
all: build

# Build for current platform
build:
	go build $(BUILD_FLAGS) -o $(BINARY_NAME) .
	chmod +x $(BINARY_NAME)

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BINARY_NAME)-windows-amd64.exe .

# Run tests
test:
	go test ./...

# Install binary to /usr/local/bin (requires sudo on Linux/macOS)
install: build
ifeq ($(UNAME_S),Linux)
	install -m 755 $(BINARY_NAME) /usr/local/bin/
else ifeq ($(UNAME_S),Darwin)
	install -m 755 $(BINARY_NAME) /usr/local/bin/
else
	@echo "On Windows, use install/windows/install.ps1 instead"
	@exit 1
endif

# Install service (user-level, no sudo required)
install-service:
ifeq ($(UNAME_S),Linux)
	mkdir -p ~/.config/systemd/user/
	cp install/linux/autotidy.service ~/.config/systemd/user/
	systemctl --user daemon-reload
	systemctl --user enable --now autotidy
else ifeq ($(UNAME_S),Darwin)
	cp install/macos/com.autotidy.daemon.plist ~/Library/LaunchAgents/
	launchctl bootstrap gui/$$(id -u) ~/Library/LaunchAgents/com.autotidy.daemon.plist
else
	@echo "On Windows, use install/windows/install.ps1 instead"
	@exit 1
endif

# Uninstall service
uninstall-service:
ifeq ($(UNAME_S),Linux)
	-systemctl --user stop autotidy
	-systemctl --user disable autotidy
	rm -f ~/.config/systemd/user/autotidy.service
	systemctl --user daemon-reload
else ifeq ($(UNAME_S),Darwin)
	-launchctl bootout gui/$$(id -u)/com.autotidy.daemon
	rm -f ~/Library/LaunchAgents/com.autotidy.daemon.plist
else
	@echo "On Windows, use install/windows/uninstall.ps1 instead"
	@exit 1
endif

# Uninstall binary
uninstall: uninstall-service
ifeq ($(UNAME_S),Linux)
	rm -f /usr/local/bin/$(BINARY_NAME)
else ifeq ($(UNAME_S),Darwin)
	rm -f /usr/local/bin/$(BINARY_NAME)
else
	@echo "On Windows, use install/windows/uninstall.ps1 instead"
	@exit 1
endif

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-*
	go clean

# Build deb package (requires nfpm)
package-deb: build
	nfpm pkg --packager deb --target $(BINARY_NAME).deb

# Build rpm package (requires nfpm)
package-rpm: build
	nfpm pkg --packager rpm --target $(BINARY_NAME).rpm

# Update Nix dependency hashes (run after changing go.mod)
nix-update-deps:
	nix run github:nix-community/gomod2nix -- generate
