ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

.SILENT:
.ONESHELL:
test:
	$(info ==== Running tests ====)
	go test ./...
