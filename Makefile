# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run
BINARY_NAME=smtp2communicator
RELEASE_VERSION_TEST_RUN=v999.999.999

all: test build

build:
	# -s removes symblols and -w debuggin info resulting in smaller binary
	# -trimpath makes sure no local paths from build time are included in the binary
	$(GOBUILD) -trimpath -o $(BINARY_NAME) -ldflags "-s -w -X main.version=${RELEASE_VERSION}" -v cmd/smtp2communicator/main.go

test:
	$(GOTEST) -v ./...

clean:
	$(GOCMD) clean
	rm -f $(BINARY_NAME)

tidy:
	$(GOCMD) mod tidy

fmt:
	@echo "if this stage fails then you need to run 'go fmt ./...' and commit again"
	@bash -c '[[ $$(go fmt ./... | wc -l) == "0" ]]'

run:
	$(GORUN) -ldflags "-s -w -X main.version=${RELEASE_VERSION_TEST_RUN}" -v cmd/smtp2communicator/main.go -c configs/smtp2communicator-dev.yaml

