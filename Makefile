# imgex - Docker Image Export Tool

BINARY_NAME=imgex
DIST_DIR=dist

.DEFAULT_GOAL := build

.PHONY: help
help:
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@echo '  build     Build the imgex binary'
	@echo '  test      Run tests'
	@echo '  clean     Clean build artifacts'
	@echo '  clib      Build C libraries'

.PHONY: build
build:
	mkdir -p $(DIST_DIR)
	go build -o $(DIST_DIR)/$(BINARY_NAME) ./cmd/imgex

.PHONY: test
test:
	go test -v ./lib

.PHONY: clean
clean:
	go clean
	rm -rf $(DIST_DIR)

.PHONY: clib
clib:
	mkdir -p $(DIST_DIR)
	go build -buildmode=c-shared -o $(DIST_DIR)/libimgex.so ./clib
	go build -buildmode=c-archive -o $(DIST_DIR)/libimgex.a ./clib