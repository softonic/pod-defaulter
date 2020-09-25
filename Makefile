BIN := pod-defaulter
CRD_OPTIONS ?= "crd:trivialVersions=true"
PKG := github.com/softonic/pod-defaulter
VERSION ?= 0.0.0-dev
ARCH ?= amd64
APP ?= pod-defaulter
NAMESPACE ?= pod-defaulter
RELEASE_NAME ?= pod-defaulter
KO_DOCKER_REPO = registry.softonic.io/pod-defaulter
REPOSITORY ?= pod-defaulter

IMAGE := $(BIN)

BUILD_IMAGE ?= golang:1.14-buster


ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: all
all: dev

.PHONY: build
build: generate
	go mod download
	GOARCH=${ARCH} go build -ldflags "-X ${PKG}/pkg/version.Version=${VERSION}" ./cmd/pod-defaulter/.../

.PHONY: test
test:
	GOARCH=${ARCH} go test -v -ldflags "-X ${PKG}/pkg/version.Version=${VERSION}" ./...

.PHONY: image
image:
	docker build -t $(IMAGE):$(VERSION) -f Dockerfile .
	docker tag $(IMAGE):$(VERSION) $(IMAGE):latest

.PHONY: dev
dev: image
	kind load docker-image $(IMAGE):$(VERSION)

.PHONY: undeploy
undeploy:
	kubectl delete -f manifest.yaml || true

.PHONY: deploy
deploy: manifest
	kubectl apply -f manifest.yaml

.PHONY: up
up: image undeploy deploy

.PHONY: docker-push
docker-push:
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: secret-values
secret-values:
	./hack/generate_helm_cert_secrets $(APP) $(NAMESPACE)

.PHONY: manifest
manifest: controller-gen helm-chart secret-values
	docker run --rm -v $(PWD):/app -w /app/ alpine/helm:3.2.3 template --release-name $(RELEASE_NAME) --set "image.tag=$(VERSION)" --set "image.repository=$(REPOSITORY)"  -f chart/pod-defaulter/values.yaml -f chart/pod-defaulter/secret.values.yaml chart/pod-defaulter > manifest.yaml

.PHONY: helm-chart
helm-chart: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) webhook paths="./..." output:crd:artifacts:config=chart/pod-defaulter/templates

.PHONY: helm-deploy
helm-deploy: helm-chart secret-values
	helm upgrade --install $(RELEASE_NAME) --namespace $(NAMESPACE) --set "image.tag=$(VERSION)" -f chart/pod-defaulter/values.yaml -f chart/pod-defaulter/secret.values.yaml chart/pod-defaulter

.PHONY: generate
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: run
run: generate fmt vet manifests
	go run ./cmd/pod-defaulter/pod-defaulter.go --tls-cert=ssl/pod-defaulter.pem --tls-key=ssl/pod-defaulter.key

.PHONY: lint
lint: vet
	[ $$(gofmt -d . | wc -l) -gt 0 ] && exit 1 || exit 0

.PHONY: manifests
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# find or download controller-gen
# download controller-gen if necessary
.PHONY: controller-gen
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
