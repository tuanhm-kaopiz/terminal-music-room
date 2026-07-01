GO ?= go
BIN_DIR := bin

.PHONY: build test lint vet clean

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/music-room ./cmd/music-room
	$(GO) build -o $(BIN_DIR)/music-roomd ./cmd/music-roomd

test:
	$(GO) test ./...

lint: vet

vet:
	$(GO) vet ./...

clean:
	rm -rf $(BIN_DIR)
