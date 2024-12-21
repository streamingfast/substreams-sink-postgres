PROJECT_NAME = substreams-sink-sql
DATE := $(shell date '+%Y-%m-%dT%H:%M:%S')
HEAD = $(shell git rev-parse HEAD)
LD_FLAGS = -X main.version=$(shell git describe --tags --always --dirty) -X main.commit=$(shell git rev-parse HEAD) -X main.date=$(DATE)
BUILD_FLAGS = -mod=readonly -ldflags='$(LD_FLAGS)'
BUILD_FOLDER = .


install:
	@echo Installing $(PROJECT_NAME)...
	@go install $(BUILD_FLAGS) ./...

uninstall:
	@echo Uninstalling $(PROJECT_NAME)...
	@rm -rf $(GOPATH)/bin/$(PROJECT_NAME)