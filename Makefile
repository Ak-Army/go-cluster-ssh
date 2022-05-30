GOLANGCI_LINT_VERSION ?= v1.42.0
GOPATH ?= $(shell ${GO_EXECUTABLE} env GOPATH)

GOLANGCI_LINT := $(GOPATH)/bin/golangci-lint
check_lint_version:
	@if [ ! "v$(shell ${GOLANGCI_LINT} version | awk '{print $$4}')" = "${GOLANGCI_LINT_VERSION}" ]; then \
		rm ${GOPATH}/bin/golangci-lint; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOPATH}/bin ${GOLANGCI_LINT_VERSION}; \
	fi
.PHONY: check_lint_version
$(GOLANGCI_LINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GOPATH}/bin ${GOLANGCI_LINT_VERSION}
#: Lint with golangci-lint
lint: $(GOLANGCI_LINT) check_lint_version
	${GOLANGCI_LINT} run \
		-c build/.golangci.yml \
		-v
	make -s ci_rmvendor
.PHONY: lint

build:
	go build -tags "gtk_3_22,pango_1_42" -o _build/go-cluster-ssh .
.PHONY: build