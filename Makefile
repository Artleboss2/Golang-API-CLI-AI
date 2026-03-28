BINARY_NAME=nim
BINARY_WINDOWS=$(BINARY_NAME).exe
MAIN=.
LDFLAGS=-ldflags="-s -w"

.PHONY: all build build-windows clean install deps

all: deps build

deps:
	@echo "Téléchargement des dépendances..."
	@go mod tidy

build: deps
	@echo "Compilation pour $(shell go env GOOS)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN)
	@echo "Binaire : ./$(BINARY_NAME)"

build-windows: deps
	@echo "Compilation pour Windows..."
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_WINDOWS) $(MAIN)
	@echo "Binaire : ./$(BINARY_WINDOWS)"

install: build
	@echo "Installation dans /usr/local/bin..."
	@sudo mv $(BINARY_NAME) /usr/local/bin/
	@echo "Installé ! Utilisez 'nim' depuis n'importe où."

clean:
	@rm -f $(BINARY_NAME) $(BINARY_WINDOWS)
	@echo "   Binaires supprimés."

test:
	@go test ./...

lint:
	@golangci-lint run ./...
