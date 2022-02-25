.PHONY: all test yamls
.FORCE:

GO_CMD := go
GO_FMT := gofmt

IMAGE_BUILD_CMD := docker build
IMAGE_BUILD_EXTRA_OPTS :=
IMAGE_PUSH_CMD := docker push

SRCROOT = $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/)
BUILD_DIR :=$(SRCROOT)/build
VERSION := $(shell git describe --tags --dirty --always)
PF9_TAG_VERSION ?= v0.6.0-pmk-$(BUILD_NUMBER)
BUILD_NUMBER ?= 1

IMAGE_REGISTRY := quay.io/kubernetes_incubator
IMAGE_NAME := node-feature-discovery
IMAGE_TAG_NAME ?= $(PF9_TAG_VERSION)
IMAGE_REPO := $(IMAGE_REGISTRY)/$(IMAGE_NAME)
IMAGE_TAG := $(IMAGE_REPO):$(IMAGE_TAG_NAME)
K8S_NAMESPACE := kube-system
HOSTMOUNT_PREFIX := /host-
KUBECONFIG :=
E2E_TEST_CONFIG :=

yaml_templates := $(wildcard *.yaml.template)
yaml_instances := $(patsubst %.yaml.template,%.yaml,$(yaml_templates))

all: image

$(BUILD_DIR):
	mkdir -p $@

image: yamls
	$(IMAGE_BUILD_CMD) --build-arg NFD_VERSION=$(VERSION) \
		--build-arg HOSTMOUNT_PREFIX=$(HOSTMOUNT_PREFIX) \
		-t $(IMAGE_TAG) \
		$(IMAGE_BUILD_EXTRA_OPTS) ./
	echo ${IMAGE_TAG} > $(BUILD_DIR)/container-tag

yamls: $(yaml_instances)

%.yaml: %.yaml.template .FORCE
	@echo "$@: namespace: ${K8S_NAMESPACE}"
	@echo "$@: image: ${IMAGE_TAG}"
	@sed -E \
	     -e s',^(\s*)name: node-feature-discovery # NFD namespace,\1name: ${K8S_NAMESPACE},' \
	     -e s',^(\s*)image:.+$$,\1image: ${IMAGE_TAG},' \
	     -e s',^(\s*)namespace:.+$$,\1namespace: ${K8S_NAMESPACE},' \
	     -e s',^(\s*)mountPath: "/host-,\1mountPath: "${HOSTMOUNT_PREFIX},' \
	     $< > $@

mock:
	mockery --name=FeatureSource --dir=source --inpkg --note="Re-generate by running 'make mock'"
	mockery --name=APIHelpers --dir=pkg/apihelper --inpkg --note="Re-generate by running 'make mock'"
	mockery --name=LabelerClient --dir=pkg/labeler --inpkg --note="Re-generate by running 'make mock'"

gofmt:
	@$(GO_FMT) -w -l $$(find . -name '*.go')

gofmt-verify:
	@out=`$(GO_FMT) -l -d $$(find . -name '*.go')`; \
	if [ -n "$$out" ]; then \
	    echo "$$out"; \
	    exit 1; \
	fi

ci-lint:
	golangci-lint run --timeout 5m0s

test:
	$(GO_CMD) test ./cmd/... ./pkg/...

e2e-test:
	$(GO_CMD) test -v ./test/e2e/ -args -nfd.repo=$(IMAGE_REPO) -nfd.tag=$(IMAGE_TAG_NAME) -kubeconfig=$(KUBECONFIG) -nfd.e2e-config=$(E2E_TEST_CONFIG)

push:
	$(IMAGE_PUSH_CMD) $(IMAGE_TAG)
