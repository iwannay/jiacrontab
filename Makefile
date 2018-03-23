# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_MANAGER=server
BINARY_CLIENT=client
BINARY_MANAGER_UNIX=$(BINARY_MANAGER)_unix
BINARY_CLIENT_UNIX=$(BINARY_CLIENT)_unix

.PHONY: all build test clean run build-linux build-windows
all: test build
build:
	$(GOBUILD) -o $(BINARY_MANAGER) -v ./server
	$(GOBUILD) -o $(BINARY_CLIENT) -v ./client
test:
	$(GOTEST) -v ./server
	$(GOTEST) -v ./client
clean:
	$(GOCLEAN)
	rm -f $(BINARY_CLIENT_UNIX)
	rm -f $(BINARY_MANAGER_UNIX)
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)


# Cross compilation
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_MANAGER)-linux-amd64 -v ./server
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_CLIENT)-linux-amd64 -v ./client

build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_MANAGER)-windows-amd64.exe -v ./server
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_CLIENT)-windows-amd64.exe -v ./client