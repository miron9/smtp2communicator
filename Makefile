# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=smtp2communicator

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v cmd/smtp2communicator/main.go

test:
	$(GOTEST) -v ./...

clean:
	$(GOCMD) clean
	rm -f $(BINARY_NAME)

