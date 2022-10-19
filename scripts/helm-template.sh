#!/bin/bash

CI_HELMFILE_REL="https://github.com/roboll/helmfile/releases/download/v0.144.0/helmfile_linux_amd64"

set -x
if [ ${CI} ]; then
    `which helmfile`
    if [ $? -ne 0 ]; then
        curl -o /usr/local/bin/helmfile -s -L ${CI_HELMFILE_REL}
        chmod +x /usr/local/bin/helmfile
    fi
fi

set -e
mkdir -p gen
helmfile -e local -f helmfile.d/01-app.yaml template > gen/deployment.yaml