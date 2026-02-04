# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

ENSURE_GARDENER_MOD         := $(shell go get github.com/gardener/gardener@$$(go list -m -f "{{.Version}}" github.com/gardener/gardener))
GARDENER_HACK_DIR           := $(shell go list -m -f "{{.Dir}}" github.com/gardener/gardener)/hack
EXTENSION_PREFIX            := gardener-extension
NAME                        := os-gardenlinux
REGISTRY                    := europe-docker.pkg.dev/gardener-project/public
IMAGE_PREFIX                := $(REGISTRY)/gardener/extensions
REPO_ROOT                   := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
HACK_DIR                    := $(REPO_ROOT)/hack
VERSION                     := $(shell cat "$(REPO_ROOT)/VERSION")
LD_FLAGS                    := "-w -X github.com/gardener/$(EXTENSION_PREFIX)-$(NAME)/pkg/version.Version=$(IMAGE_TAG)"
LEADER_ELECTION             := true
IGNORE_OPERATION_ANNOTATION := true
PLATFORM                    ?= linux/amd64

#########################################
# Tools                                 #
#########################################

TOOLS_DIR := $(HACK_DIR)/tools
include $(GARDENER_HACK_DIR)/tools.mk

#########################################
# Rules for local development scenarios #
#########################################

.PHONY: start
start:
	@LEADER_ELECTION_NAMESPACE=default GO111MODULE=on go run \
		-ldflags $(LD_FLAGS) \
		./cmd/$(EXTENSION_PREFIX)-$(NAME) \
		--leader-election=$(LEADER_ELECTION) \
		--ignore-operation-annotation=$(IGNORE_OPERATION_ANNOTATION) \
		--gardener-version="v1.56.0"

#################################################################
# Rules related to binary build, Docker image build and release #
#################################################################

.PHONY: install
install:
	@LD_FLAGS="-w -X github.com/gardener/$(EXTENSION_PREFIX)-$(NAME)/pkg/version.Version=$(VERSION)" \
	bash $(GARDENER_HACK_DIR)/install.sh ./...

.PHONY: docker-login
docker-login:
	@gcloud auth activate-service-account --key-file .kube-secrets/gcr/gcr-readwrite.json

.PHONY: docker-images
docker-images:
	@docker buildx build --platform=$(PLATFORM) \
		-t $(IMAGE_PREFIX)/$(NAME):$(VERSION) \
		-t $(IMAGE_PREFIX)/$(NAME):latest \
		-f Dockerfile \
		-m 6g \
		--target $(EXTENSION_PREFIX)-$(NAME) \
		.

#####################################################################
# Rules for verification, formatting, linting, testing and cleaning #
#####################################################################

.PHONY: tidy
tidy:
	@GO111MODULE=on go mod tidy
	@mkdir -p $(REPO_ROOT)/.ci/hack && cp $(GARDENER_HACK_DIR)/.ci/* $(REPO_ROOT)/.ci/hack/ && chmod +xw $(REPO_ROOT)/.ci/hack/*
	@GARDENER_HACK_DIR=$(GARDENER_HACK_DIR) $(REPO_ROOT)/hack/update-github-templates.sh
	@cp $(GARDENER_HACK_DIR)/cherry-pick-pull.sh $(HACK_DIR)/cherry-pick-pull.sh && chmod +xw $(HACK_DIR)/cherry-pick-pull.sh

.PHONY: clean
clean:
	@$(shell find ./example -type f -name "controller-registration.yaml" -exec rm '{}' \;)
	@bash $(GARDENER_HACK_DIR)/clean.sh ./cmd/... ./pkg/...

.PHONY: check-generate
check-generate:
	@bash $(GARDENER_HACK_DIR)/check-generate.sh $(REPO_ROOT)

.PHONY: check
check: $(GOIMPORTS) $(GOLANGCI_LINT)
	@REPO_ROOT=$(REPO_ROOT) bash $(GARDENER_HACK_DIR)/check.sh --golangci-lint-config=./.golangci.yaml ./cmd/... ./pkg/...
	@REPO_ROOT=$(REPO_ROOT) bash $(GARDENER_HACK_DIR)/check-charts.sh ./charts

.PHONY: generate
generate: $(CONTROLLER_GEN) $(EXTENSION_GEN) $(GEN_CRD_API_REFERENCE_DOCS) $(HELM) $(KUSTOMIZE) $(MOCKGEN) $(YQ)
	@REPO_ROOT=$(REPO_ROOT) GARDENER_HACK_DIR=$(GARDENER_HACK_DIR) bash $(GARDENER_HACK_DIR)/generate-sequential.sh ./charts/... ./cmd/... ./example/... ./pkg/...
	$(MAKE) format

.PHONY: add-license-headers
add-license-headers: $(GO_ADD_LICENSE)
	@REPO_ROOT=$(REPO_ROOT) bash $(GARDENER_HACK_DIR)/add-license-header.sh

.PHONY: check-license-headers
check-license-headers: $(GO_ADD_LICENSE)
	@REPO_ROOT=$(REPO_ROOT) bash $(GARDENER_HACK_DIR)/check-license-header.sh

.PHONY: sast
sast: $(GOSEC)
	@./hack/sast.sh

.PHONY: sast-report
sast-report: $(GOSEC)
	@./hack/sast.sh --gosec-report true

.PHONY: format
format: $(GOIMPORTS) $(GOIMPORTSREVISER)
	@bash $(GARDENER_HACK_DIR)/format.sh ./cmd ./pkg

.PHONY: test
test:
	@bash $(GARDENER_HACK_DIR)/test.sh ./cmd/... ./pkg/...

.PHONY: test-cov
test-cov:
	@bash $(GARDENER_HACK_DIR)/test-cover.sh ./cmd/... ./pkg/...

.PHONY: test-clean
test-clean:
	@bash $(GARDENER_HACK_DIR)/test-cover-clean.sh

.PHONY: verify
verify: check format sast test check-license-headers

.PHONY: verify-extended
verify-extended: check-generate check format sast-report test-cov test-clean check-license-headers
