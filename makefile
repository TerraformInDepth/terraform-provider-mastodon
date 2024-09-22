default: build

.PHONY: build
build:
	go build -v ./...

.PHONY: install
install: build
	go install -v ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: fmt
fmt:
	gofmt -s -w -e .

.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

.PHONY: docs
docs:
	@echo "Generating docs"
	go generate

.PHONY: update_deps
update_deps:
	@echo "Updating dependencies"
	go get -u ./...
	go mod tidy
