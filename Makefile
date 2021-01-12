BINARIES ?= chimera-admission-amd64
BUILD_DIR := build
GOLANGCI_LINT_VERSION = v1.35.2

.PHONY: phony-explicit

.PHONY: chimera-admission
chimera-admission: $(BINARIES)

chimera-admission-%: phony-explicit
	sh -c 'CGO_ENABLED=1 GOOS=linux GOARCH=$* GO111MODULE=on go build -o chimera-admission-$*'

.PHONY: run
run:
	sh -c 'GO111MODULE=on go run main.go'

.PHONY: clean
clean:
	rm -f $(BINARIES)

.PHONY: verify
verify: verify-go-lint

.PHONY: verify-go-lint
verify-go-lint: $(BUILD_DIR)/golangci-lint
	$(BUILD_DIR)/golangci-lint run --timeout=2m

$(BUILD_DIR)/golangci-lint:
	export \
		VERSION=$(GOLANGCI_LINT_VERSION) \
		URL=https://raw.githubusercontent.com/golangci/golangci-lint \
		BINDIR=$(BUILD_DIR) && \
	curl -sfL $$URL/$$VERSION/install.sh | sh -s $$VERSION
	$(BUILD_DIR)/golangci-lint version
	$(BUILD_DIR)/golangci-lint linters
