# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean

# Build directories
SERVER_DIR=cmd/server
CLIENT_DIR=cmd/client

# Binary names
SERVER_BINARY=server
CLIENT_BINARY=client

# Default port
PORT=8080

all: build

build: build-server build-client

build-server:
	cd $(SERVER_DIR) && $(GOBUILD) -o $(SERVER_BINARY)

build-client:
	cd $(CLIENT_DIR) && $(GOBUILD) -o $(CLIENT_BINARY)

run-server:
	cd $(SERVER_DIR) && ./$(SERVER_BINARY) -port $(PORT)

run-client:
	cd $(CLIENT_DIR) && ./$(CLIENT_BINARY) -server localhost:$(PORT)

clean:
	cd $(SERVER_DIR) && $(GOCLEAN)
	cd $(CLIENT_DIR) && $(GOCLEAN)
	rm -f $(SERVER_DIR)/$(SERVER_BINARY)
	rm -f $(CLIENT_DIR)/$(CLIENT_BINARY)

.PHONY: all build build-server build-client run-server run-client clean