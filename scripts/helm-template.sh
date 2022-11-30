#!/bin/bash

CI_HELMFILE_REL="https://github.com/roboll/helmfile/releases/download/v0.144.0/helmfile_linux_amd64"
CI_HELMFILE_SHA="dcf865a715028d3a61e2fec09f2a0beaeb7ff10cde32e096bf94aeb9a6eb4f02"
CI_KUBECONFORM_REL="https://github.com/yannh/kubeconform/releases/download/v0.5.0/kubeconform-linux-amd64.tar.gz"
CI_KUBECONFORM_SHA="5b39700d0924072313ad7e898b6101ea0ebdd3634301b1176b25a8572e62190e"

set -x
if [ ${CI} ]; then
    `which helmfile`
    if [ $? -ne 0 ]; then
        curl -o /usr/local/bin/helmfile -s -L ${CI_HELMFILE_REL}
        chmod +x /usr/local/bin/helmfile
        sha256sum /usr/local/bin/helmfile | grep -q ${CI_HELMFILE_SHA}
        if [ $? -ne 0 ]; then
            echo "ERROR: invalid sha256 for helmfile"
            exit 1
        fi        
    fi
    `which kubeconform`
    if [ $? -ne 0 ]; then
        curl -o /tmp/kubeconform.tar.gz -s -L ${CI_KUBECONFORM_REL}
        sha256sum /tmp/kubeconform.tar.gz | grep -q ${CI_KUBECONFORM_SHA}
        if [ $? -ne 0 ]; then
            echo "ERROR: invalid sha256 for kubeconform"
            exit 1
        fi
        tar -xvf /tmp/kubeconform.tar.gz -C /usr/local/bin/
        chmod +x /usr/local/bin/kubeconform
    fi
fi

set -e
mkdir -p gen
helmfile -e local -f helmfile.d/01-app.yaml template > gen/deployment.yaml
kubeconform -kubernetes-version 1.24.0 -strict -verbose -summary ./gen