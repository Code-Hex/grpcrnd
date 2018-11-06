$(shell go env GOPATH)/bin/dep:
	@go get -u -v github.com/golang/dep/cmd/dep

.PHONY: dep
dep: $(shell go env GOPATH)/bin/dep
	@echo "+ $@"
	@dep ensure -v

.PHONY: dep/vendor-only
dep/vendor-only: $(shell go env GOPATH)/bin/dep
	@echo "+ $@"
	@dep ensure -v -vendor-only

.PHONY: vet
vet:
	@go vet ./...

.PHONY: build
build:
	@echo "+ $@"
	CGO_ENABLED=0 go build -o bin/server \
        -ldflags "-w -s" \
        github.com/Code-Hex/grpcrnd/cmd/grpcrnd


.PHONY: test
test:
	@echo "+ $@"
	@go test -v -race ./...

.PHONY: help
help:
	@perl -nle 'BEGIN {printf "Usage:\n  make \033[33m<target>\033[0m\n\nTargets:\n"} printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 if /^([a-zA-Z_-].+):\s+## (.*)/' $(MAKEFILE_LIST)