#!/bin/bash
set -e

K8S_RESTDEV_OPERATOR=$1
PROJECT_ID=$2
CONTROLLER=$K8S_RESTDEV_OPERATOR-controller.go
BIN=bin/k8s-restdev-$K8S_RESTDEV_OPERATOR
GOOS=linux go build \
  -a --ldflags '-extldflags "-static"' \
  -tags netgo \
  -installsuffix netgo \
  -v \
  -o $BIN $CONTROLLER

IMAGE=gcr.io/$PROJECT_ID/k8s-restdev-$K8S_RESTDEV_OPERATOR:kube-system.$(date +%s)
docker build --build-arg bin=$BIN  -t $IMAGE .
