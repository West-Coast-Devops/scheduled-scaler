DATE:=$(shell date +%s)

GOBIN=go
GOBUILD=$(GOBIN) build
GOTEST=$(GOBIN) test

CONTROLLER=$(OPERATOR)-controller.go
BIN=bin/k8s-restdev-$(OPERATOR)

DOCKERBIN=docker
IMAGE?=gcr.io/$(PROJECT_ID)/k8s-restdev-$(OPERATOR):kube-system.$(DATE)
DOCKERBUILD=$(DOCKERBIN) build --build-arg bin=$(BIN) -t $(IMAGE) .

DEPLOYBIN?=kubectl
KN_PROJECT_ID?=$(PROJECT_ID)

.PHONY: test

all: test build push deploy
test:
	$(GOTEST) ./test/$(OPERATOR)
build:
	GOOS=linux $(GOBUILD) \
  -a --ldflags '-extldflags "-static"' \
  -tags netgo \
  -installsuffix netgo \
  -v \
  -o $(BIN) $(CONTROLLER)
	$(DOCKERBUILD)
push:
	gcloud docker -- push $(IMAGE)
deploy:
ifeq ($(DEPLOYBIN), kn)
	cat ./artifacts/kubes/$(OPERATOR)/deployment.yml | sed "s|\[IMAGE\]|$(IMAGE)|g" | kn $(KN_PROJECT_ID) -- --namespace=kube-system apply -f -
else
	cat ./artifcats/kubes/$(OPERATOR)/deployment.yml | sed "s|\[IMAGE\]|$(IMAGE)|g" | kubectl --namespace=kube-system apply -f -
endif
