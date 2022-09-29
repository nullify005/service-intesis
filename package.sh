#!/bin/bash

set -ex

get_abs_filename() {
  # $1 : relative filename
  echo "$(cd "$(dirname "$1")" && pwd)/$(basename "$1")"
}

clean() {
    rm -rf ${1}
    mkdir -p ${1}
}

IMG="nullify005/service-intesis:development"
BUILDX="docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 -t ${IMG}"
ROOT_DIR=$(dirname `get_abs_filename $0`)
TRIVY_OPTS="--ignorefile ${ROOT_DIR}/.trivyignore --severity CRITICAL,HIGH,MEDIUM --exit-code 1"
BUILD_DIR=${ROOT_DIR}/tmp

# conduct a build & test
clean ${BUILD_DIR}
trivy fs ${TRIVY_OPTS} .
${BUILDX} -o type=oci,dest=${BUILD_DIR}/image.tar .    
(
    cd ${BUILD_DIR}
    tar -xvf image.tar 
    rm image.tar
    trivy i ${TRIVY_OPTS} --ignore-unfixed --input .
)
clean ${BUILD_DIR}
# helmfile -e development -f helmfile.d/01-app.yaml template > ${BUILD_DIR}/deploy.yaml
# (
#     cd ${BUILD_DIR}
#     trivy config ${TRIVY_OPTS} .
# )
# rm -rf ${BUILD_DIR}

# assuming we're successful, now push to the repo
docker login
${BUILDX} -o type=registry .