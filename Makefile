.PHONY: build clean dev version demo $(TARGETS)

BINARY_NAME := convit
VERSION := 0.5.0
BUILD_DIR := bin

TARGETS := darwin-arm64 darwin-amd64 linux-arm64 linux-amd64
LDFLAGS := -w -s -X main.AppVersion=$(VERSION) -X main.AppName=$(BINARY_NAME)

build: $(TARGETS)

$(TARGETS):
	GOOS=$(word 1,$(subst -, ,$@)) GOARCH=$(word 2,$(subst -, ,$@)) go build -o $(BUILD_DIR)/$(BINARY_NAME)-$@ -ldflags "$(LDFLAGS)"

clean:
	rm -rf $(BUILD_DIR)

dev:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) -ldflags "$(LDFLAGS)"

version:
	@echo $(VERSION)

lint:
	@go fmt .

demo:
	@vhs demo.tape
