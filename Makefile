DATE:=$(shell date +%s)

GOBIN=go
GOBUILD=$(GOBIN) build
GOTEST=$(GOBIN) test
GOPATH ?= $(shell go env GOPATH)

OPERATOR?=scaling
CONTROLLER=$(OPERATOR)-controller.go
TEST_CONTROLLER=$(OPERATOR)-controller_test.go
BIN=bin/k8s-restdev-$(OPERATOR)

DOCKERBIN=docker
VERSION?=kube-system.$(DATE)
IMAGE?=k8srestdev/$(OPERATOR):$(VERSION)
DOCKERBUILD=$(DOCKERBIN) build --build-arg bin=$(BIN) -t $(IMAGE) .

DEPLOYBIN?=kubectl
KN_PROJECT_ID?=$(PROJECT_ID)

.PHONY: test

all: codegen test build push deploy
ci: codegen test build push
localbin: codegen test build
codegen:
	./hack/update-codegen.sh && \
	go generate ./...
test:
	$(GOTEST) $(TEST_CONTROLLER) $(CONTROLLER)
	$(GOTEST) ./...
build:
	GOOS=linux $(GOBUILD) \
  -a --ldflags '-extldflags "-static"' \
  -tags netgo \
  -installsuffix netgo \
  -v \
  -o $(BIN) $(CONTROLLER)
	$(DOCKERBUILD)
push:
	docker push $(IMAGE)
deploy:
ifeq ($(DEPLOYBIN), kn)
	cat ./artifacts/kubes/$(OPERATOR)/deployment.yml | sed "s|\[IMAGE\]|$(IMAGE)|g" | kn $(KN_PROJECT_ID) -- --namespace=kube-system apply -f -
else
	cat ./artifacts/kubes/$(OPERATOR)/deployment.yml | sed "s|\[IMAGE\]|$(IMAGE)|g" | kubectl --namespace=kube-system apply -f -
endif
