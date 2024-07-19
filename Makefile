.PHONY: build clean run version demo

BINARY_NAME=convit
VERSION=0.1.1
BUILD_DIR=bin

TARGETS=darwin-arm64 darwin-amd64 linux-arm64 linux-amd64
LDFLAGS="-w -s -X main.AppVersion=$(VERSION) -X main.AppName=$(BINARY_NAME)"

build: $(TARGETS)

darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 -ldflags $(LDFLAGS)

darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 -ldflags $(LDFLAGS)

linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 -ldflags $(LDFLAGS)

linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 -ldflags $(LDFLAGS)

clean:
	rm -rf $(BUILD_DIR)

dev:
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) -ldflags $(LDFLAGS)

version:
	@echo $(VERSION)

lint:
	@go fmt .

demo:
	@vhs demo.tape
