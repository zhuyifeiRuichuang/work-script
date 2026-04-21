# Image URL to use all building/pushing image targets
# 默认镜像名称与部署配置 config/manager/manager.yaml 保持一致
IMG ?= apache/hadoop-operator:latest
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.29.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are a selection of the "# Category"
# from the beginning of the target's comment block. If a target doesn't have
# the "# Category" comment block, then it is categorized as "General".
.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/main.go

# If you wish built the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64 ). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: test ## Build docker image with the manager.
	$(CONTAINER_TOOL) build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push ${IMG}

# PLATFORMS defines the target platforms for  the manager image be build to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - able to use docker buildx . More info: https://docs.docker.com/build/buildx/
# - have enable BuildKit, More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image for your registry (i.e. if you do not inform a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To properly provided solutions that supports more than one platform you should use this option.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le

# Hadoop 镜像相关变量
HADOOP_VERSION ?= 3.3.6
HADOOP_IMG ?= apache/hadoop:$(HADOOP_VERSION)

# 私有仓库配置（用于推送镜像）
DOCKER_REGISTRY ?=

# 计算最终镜像名称（如果有 DOCKER_REGISTRY 则添加前缀）
ifeq ($(DOCKER_REGISTRY),)
    HADOOP_TARGET_IMG = $(HADOOP_IMG)
else
    HADOOP_TARGET_IMG = $(DOCKER_REGISTRY)/$(HADOOP_IMG)
endif

# 构建目录
BUILD_DIR ?= build

.PHONY: docker-buildx
docker-buildx: test ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$\$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$\$\{BUILDPLATFORM\}/' $(BUILD_DIR)/operator/Dockerfile > Dockerfile.cross
	- $(CONTAINER_TOOL) buildx create --name project-v3-builder
	$(CONTAINER_TOOL) buildx use project-v3-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- $(CONTAINER_TOOL) buildx rm project-v3-builder
	rm Dockerfile.cross

##@ Hadoop Images

.PHONY: build-images
build-images: build-operator-image build-hadoop-image ## Build all images (Operator and Hadoop)
	@echo "All images built successfully!"
	@echo "Operator Image: $(IMG)"
	@echo "Hadoop Image: $(HADOOP_TARGET_IMG)"

.PHONY: build-operator-image
build-operator-image: ## Build only Operator image
	@echo "Building Operator image: $(IMG)"
	$(CONTAINER_TOOL) build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ') \
		--build-arg GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown') \
		--build-arg GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown') \
		-t $(IMG) \
		-f $(BUILD_DIR)/operator/Dockerfile .

.PHONY: build-hadoop-image
build-hadoop-image: ## Build only Hadoop cluster image
	@echo "Building Hadoop image: $(HADOOP_TARGET_IMG)"
	$(CONTAINER_TOOL) build \
		--build-arg HADOOP_VERSION=$(HADOOP_VERSION) \
		-t $(HADOOP_TARGET_IMG) \
		-f $(BUILD_DIR)/hadoop/Dockerfile $(BUILD_DIR)/hadoop/

.PHONY: push-images
push-images: ## Push all images to registry
	@echo "Pushing images to registry..."
	@echo "Pushing Operator: $(IMG)"
	$(CONTAINER_TOOL) push $(IMG)
	@echo "Pushing Hadoop: $(HADOOP_TARGET_IMG)"
	$(CONTAINER_TOOL) push $(HADOOP_TARGET_IMG)

.PHONY: build-hadoop-multiarch
build-hadoop-multiarch: ## Build multi-arch Hadoop image
	@echo "Building multi-arch Hadoop image: $(HADOOP_TARGET_IMG)"
	$(CONTAINER_TOOL) buildx build \
		--platform $(PLATFORMS) \
		--build-arg HADOOP_VERSION=$(HADOOP_VERSION) \
		-t $(HADOOP_TARGET_IMG) \
		-f $(BUILD_DIR)/hadoop/Dockerfile \
		--push $(BUILD_DIR)/hadoop/

##@ Image Utilities

.PHONY: image-lint
image-lint: ## Lint Dockerfile using hadolint (if installed)
	@if command -v hadolint >/dev/null 2>&1; then \
		echo "Linting Dockerfiles..."; \
		hadolint $(BUILD_DIR)/operator/Dockerfile; \
		hadolint $(BUILD_DIR)/hadoop/Dockerfile; \
	else \
		echo "hadolint not installed, skipping linting"; \
		echo "Install: https://github.com/hadolint/hadolint"; \
	fi

.PHONY: image-scan
image-scan: docker-build ## Scan Operator image for vulnerabilities using trivy (if installed)
	@if command -v trivy >/dev/null 2>&1; then \
		echo "Scanning Operator image..."; \
		trivy image $(IMG); \
	else \
		echo "trivy not installed, skipping scan"; \
		echo "Install: https://github.com/aquasecurity/trivy"; \
	fi

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
KUSTOMIZE_VERSION ?= v5.0.1
CONTROLLER_TOOLS_VERSION ?= v0.14.0

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary (ideally with version)
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f $(1) ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv "$$(echo "$(1)" | sed "s/-$(3)$$//")" "$(1)" ;\
}
endef
