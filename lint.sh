#!/bin/bash

set -ex

get_abs_filename() {
  # $1 : relative filename
  echo "$(cd "$(dirname "$1")" && pwd)/$(basename "$1")"
}

IMG="nullify005/service-intesis:development"
ROOT_DIR=$(dirname `get_abs_filename $0`)
TRIVY_OPTS="--ignorefile ${ROOT_DIR}/.trivyignore --severity CRITICAL,HIGH,MEDIUM --exit-code 1"

# conduct a build & test
docker build --target test .
docker build -t ${IMG} .
./scripts/helm-template.sh
trivy fs ${TRIVY_OPTS} .
trivy config ${TRIVY_OPTS} .
trivy i ${TRIVY_OPTS} --ignore-unfixed ${IMG}