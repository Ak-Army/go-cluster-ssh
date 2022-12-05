GOLANGCI_LINT_VERSION ?= v1.42.0
GOPATH ?= $(shell ${GO_EXECUTABLE} env GOPATH)
PACKAGE_NAME ?= $(shell go mod edit -json | grep 'Path' | head -1 | sed -re 's/.*: "([^"]+)"/\1/')

BUILD_VERSION ?= $(shell git describe --tags)
BUILD_TIME = $(shell date +%FT%T%z)

LD_FLAGS = "-X "${PACKAGE_NAME}/internal".Version=${BUILD_VERSION} \
	-X "${PACKAGE_NAME}/internal".BuildTime=${BUILD_TIME}"

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
	go build \
		-v \
		-tags "gtk_3_22,pango_1_42" \
		-o _build/go-cluster-ssh \
		-ldflags=${LD_FLAGS} \
		.
.PHONY: build

#: Build docker builder image
docker-builder-image:
	docker image build \
		--network host \
		-t go-cluster-ssh:dev \
		.
.PHONY: docker-builder-image

docker-build: docker-builder-image
	docker run \
 	--rm --network host \
	--user $$(id -u):$$(id -g) \
	-v /etc/group:/etc/group:ro \
	-v /etc/passwd:/etc/passwd:ro \
	-v /etc/shadow:/etc/shadow:ro \
	-v ${HOME}/.cache:${HOME}/.cache \
	-v $(shell pwd):/build \
	-v ${GOPATH}:/go \
	-e HOME=${HOME} \
	go-cluster-ssh:dev \
	make build
.PHONY: docker-build