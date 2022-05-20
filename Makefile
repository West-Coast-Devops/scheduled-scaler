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

.PHONY: test all ci localbin tools

all: tools codegen test build push deploy
ci: tools codegen test build push
localbin: tools codegen test build
tools:
	go run k8s.restdev.com/operators/tools/gettools -v 2 -alsologtostderr
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
