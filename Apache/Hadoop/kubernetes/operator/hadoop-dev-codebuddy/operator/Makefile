# VERSION defines the version of the operator
VERSION ?= 0.1.0
# IMG_TAG defines the tag for the operator image
IMG_TAG ?= $(VERSION)
# REGISTRY defines the registry to push images
REGISTRY ?= quay.io/kubedoop
# IMAGE_NAME defines the name of the operator image
IMAGE_NAME ?= hadoop-operator
# IMG defines the full image name
IMG ?= $(REGISTRY)/$(IMAGE_NAME):$(IMG_TAG)
# GOFLAGS defines additional go flags
GOFLAGS ?=
# ENVTEST_K8S_VERSION defines the version of Kubernetes for envtest
ENVTEST_K8S_VERSION ?= 1.29.0
# KUSTOMIZE defines the kustomize binary to use
KUSTOMIZE ?= $(LOCALBIN)/kustomize
# CONTROLLER_GEN defines the controller-gen binary to use
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
# GOLANGCI_LINT defines the golangci-lint binary to use
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
# HELM defines the helm binary to use
HELM ?= $(LOCALBIN)/helm
# CHAINSAW defines the chainsaw binary to use
CHAINSAW ?= $(LOCALBIN)/chainsaw
# KIND defines the kind binary to use
KIND ?= $(LOCALBIN)/kind
# KUBECTL defines the kubectl binary to use
KUBECTL ?= $(LOCALBIN)/kubectl
# GINKGO defines the ginkgo binary to use
GINKGO ?= $(LOCALBIN)/ginkgo

# Local bin directory
LOCALBIN := $(shell pwd)/bin
$(LOCALBIN) %:
	@mkdir -p $(LOCALBIN)

# ===========================================
# Go commands
# ===========================================

# Run go vet
.PHONY: vet
vet:
	go vet ./...

# Run go fmt
.PHONY: fmt
fmt:
	go fmt ./...

# Run golangci-lint
.PHONY: lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run

$(GOLANGCI_LINT): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.0

# ===========================================
# Build commands
# ===========================================

# Build manager binary
.PHONY: build
build: fmt vet
	go build -o bin/manager $(GOFLAGS) cmd/main.go

# Build all binaries
.PHONY: build.all
build.all: build

# Run tests
.PHONY: test
test: manifests generate fmt vet
	go test ./... -v -count=1

# Run unit tests with coverage
.PHONY: test-coverage
test-coverage:
	go test ./... -v -count=1 -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Run e2e tests
.PHONY: test-e2e
test-e2e: $(KIND) $(CHAINSAW) $(KUBECTL)
	$(KIND) create cluster --name=hadoop-operator-test || true
	$(KIND) load docker-image $(IMG) --name=hadoop-operator-test || true
	$(KUBECTL) apply -f https://quay.io/kubedoop/commons-operator?ref=main || true
	$(KUBECTL) apply -f https://quay.io/kubedoop/zookeeper-operator?ref=main || true
	$(KUBECTL) apply -f deploy/helm/hadoop-operator/crds/
	$(KUBECTL) apply -f deploy/helm/hadoop-operator/
	$(CHAINSAW) test --test-dir=test/e2e --report-name=hadoop-operator-e2e-report
	$(KIND) delete cluster --name=hadoop-operator-test

$(KIND) $(CHAINSAW) $(KUBECTL): $(LOCALBIN)
	@echo "Installing binaries..."

# ===========================================
# Generate commands
# ===========================================

# Generate code (DeepCopy, etc.)
.PHONY: generate
generate: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Generate manifests (CRDs, RBAC, etc.)
.PHONY: manifests
manifests: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) crd:allowGenerationCircleSkipped=true paths="./..." output:crd:artifacts:config=config/crd/bases
	$(CONTROLLER_GEN) rbac:roleName=manager-role paths="./..." output:rbac:artifacts:config=config/rbac

$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0

# ===========================================
# Docker commands
# ===========================================

# Build the operator image
.PHONY: docker-build
docker-build:
	docker build -t $(IMG) -f Dockerfile .

# Build the operator image with no cache
.PHONY: docker-build-no-cache
docker-build-no-cache:
	docker build --no-cache -t $(IMG) -f Dockerfile .

# Push the operator image
.PHONY: docker-push
docker-push:
	docker push $(IMG)

# Build and push multi-platform images
.PHONY: docker-buildx
docker-buildx:
	docker buildx build --platform linux/amd64,linux/arm64 -t $(IMG) -f Dockerfile . --push

# Build and load local image for kind testing
.PHONY: docker-build-kind
docker-build-kind: docker-build
	$(KIND) load docker-image $(IMG) --name=hadoop-operator || true

# ===========================================
# Kustomize commands
# ===========================================

# Install kustomize if not present
$(KUSTOMIZE): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v5@v5.4.0

# Generate kustomized deployment
.PHONY: kustomize-build
kustomize-build: $(KUSTOMIZE)
	$(KUSTOMIZE) build config/default > deploy/manifests.yaml

# ===========================================
# Helm commands
# ===========================================

# Install helm if not present
$(HELM): $(LOCALBIN)
	@echo "Installing helm..."
	@if command -v brew >/dev/null 2>&1; then \
		brew install helm; \
	elif command -v apt-get >/dev/null 2>&1; then \
		curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash; \
	else \
		curl -fsSL https://get.helm.sh/helm-v3.14.0-linux-amd64.tar.gz -o /tmp/helm.tar.gz && \
		tar -zxf /tmp/helm.tar.gz -C /tmp && \
		mv /tmp/linux-amd64/helm $(LOCALBIN)/helm && \
		rm -rf /tmp/helm.tar.gz /tmp/linux-amd64; \
	fi
	@echo "Helm installed successfully"

# Lint helm chart
.PHONY: helm-lint
helm-lint: $(HELM)
	$(HELM) lint deploy/helm/hadoop-operator

# Template helm chart
.PHONY: helm-template
helm-template: $(HELM)
	$(HELM) template my-release deploy/helm/hadoop-operator

# Install CRDs from helm chart
.PHONY: helm-crd-sync
helm-crd-sync: $(HELM)
	@echo "Copying CRDs to Helm chart..."
	@mkdir -p deploy/helm/hadoop-operator/crds
	@cp config/crd/bases/*.yaml deploy/helm/hadoop-operator/crds/

# Package helm chart
.PHONY: helm-chart-package
helm-chart-package: $(HELM) helm-crd-sync
	$(HELM) package deploy/helm/hadoop-operator -d dist/

# Publish helm chart to OCI registry
.PHONY: helm-chart-publish
helm-chart-publish:
	$(HELM) push deploy/helm/hadoop-operator-$(VERSION).tgz oci://$(REGISTRY)/charts

# ===========================================
# Deployment commands
# ===========================================

# Install CRDs to the cluster
.PHONY: install
install: $(KUSTOMIZE)
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

# Uninstall CRDs from the cluster
.PHONY: uninstall
uninstall: $(KUSTOMIZE)
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete -f - --ignore-not-found=true

# Deploy operator to the cluster
.PHONY: deploy
deploy: $(KUSTOMIZE)
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

# Undeploy operator from the cluster
.PHONY: undeploy
undeploy: $(KUSTOMIZE)
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete -f - --ignore-not-found=true

# Deploy operator with Helm
.PHONY: helm-deploy
helm-deploy: $(HELM)
	$(HELM) upgrade --install hadoop-operator deploy/helm/hadoop-operator \
		--namespace hadoop-operator-system \
		--create-namespace \
		--set image.repository=$(REGISTRY) \
		--set image.tag=$(IMG_TAG)

# Uninstall operator deployed with Helm
.PHONY: helm-undeploy
helm-undeploy: $(HELM)
	$(HELM) uninstall hadoop-operator --namespace hadoop-operator-system || true

# ===========================================
# Development commands
# ===========================================

# Create a local development cluster with kind
.PHONY: kind-create
kind-create: $(KIND)
	$(KIND) create cluster --name=hadoop-operator-dev || true

# Delete the local development cluster
.PHONY: kind-delete
kind-delete: $(KIND)
	$(KIND) delete cluster --name=hadoop-operator-dev || true

# Set up the development environment
.PHONY: envtest
envtest: $(KIND) $(KUBECTL)
	$(KIND) create cluster --name=hadoop-operator-dev || true
	$(KUBECTL) apply -f config/crd/bases/
	$(KUBECTL) apply -f config/rbac/
	$(KUBECTL) apply -f config/manager/

# Clean up the development environment
.PHONY: envtest-clean
envtest-clean: $(KIND) $(KUBECTL)
	$(KUBECTL) delete -f config/crd/bases/ --ignore-not-found=true || true
	$(KUBECTL) delete -f config/rbac/ --ignore-not-found=true || true
	$(KUBECTL) delete -f config/manager/ --ignore-not-found=true || true
	$(KIND) delete cluster --name=hadoop-operator-dev || true

# ===========================================
# Help
# ===========================================

# Show this help
.PHONY: help
help:
	@echo "Hadoop Kubernetes Operator Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@echo ""
	@echo "Development:"
	@echo "  fmt               Run go fmt"
	@echo "  vet               Run go vet"
	@echo "  lint              Run golangci-lint"
	@echo "  test              Run unit tests"
	@echo "  test-coverage     Run unit tests with coverage"
	@echo "  test-e2e          Run e2e tests"
	@echo ""
	@echo "Build:"
	@echo "  build             Build manager binary"
	@echo "  docker-build      Build Docker image"
	@echo "  docker-push       Push Docker image"
	@echo "  docker-buildx     Build multi-platform Docker images"
	@echo ""
	@echo "Generate:"
	@echo "  generate          Generate DeepCopy code"
	@echo "  manifests         Generate CRDs and RBAC manifests"
	@echo ""
	@echo "Deploy:"
	@echo "  install           Install CRDs to cluster"
	@echo "  uninstall         Uninstall CRDs from cluster"
	@echo "  deploy            Deploy operator to cluster"
	@echo "  undeploy          Undeploy operator from cluster"
	@echo "  helm-deploy       Deploy operator using Helm"
	@echo "  helm-undeploy     Uninstall operator deployed with Helm"
	@echo ""
	@echo "Helm:"
	@echo "  helm-lint         Lint Helm chart"
	@echo "  helm-template     Template Helm chart"
	@echo "  helm-crd-sync     Sync CRDs to Helm chart"
	@echo "  helm-chart-package    Package Helm chart"
	@echo "  helm-chart-publish    Publish Helm chart to OCI registry"
	@echo ""
	@echo "Environment:"
	@echo "  VERSION           Operator version (default: 0.1.0)"
	@echo "  IMG_TAG           Image tag (default: VERSION)"
	@echo "  REGISTRY          Image registry (default: quay.io/kubedoop)"
	@echo "  IMG               Full image name (default: REGISTRY/IMAGE_NAME:TAG)"
	@echo ""
	@echo "Examples:"
	@echo "  make build VERSION=1.0.0"
	@echo "  make docker-build IMG_TAG=latest REGISTRY=my-registry"
	@echo "  make helm-deploy VERSION=0.2.0"
