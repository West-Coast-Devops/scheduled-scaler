#!/bin/bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG="../../k8s.io/code-generator"
MODULE_NAME="k8s.restdev.com/operators"
MODULE_PATH="${GOPATH}/src/${MODULE_NAME}"
if [ $PWD != $MODULE_PATH ]; then
  echo "Invalid module path! Please refer to the documentation..."
  echo -e "${RED}Current Path:${NC} $PWD"
  echo -e "${GREEN}Required Path:${NC} $MODULE_PATH"
  exit 1
fi

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
${CODEGEN_PKG}/generate-groups.sh "deepcopy,client,informer,lister" \
  ${MODULE_NAME}/pkg/client ${MODULE_NAME}/pkg/apis \
  scaling:v1alpha1

# To use your own boilerplate text append:
#   --go-header-file ${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt
