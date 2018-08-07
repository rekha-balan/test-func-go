#!/usr/bin/env bash

## prolog
set -o errexit
__filename=${BASH_SOURCE[0]}
__dirname=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "${__dirname}/../.env"
## end prolog

## parameters
declare -i publish=${1:-0}  # 0: false; 1: true
declare run_image_uri=${4:-"${RUNTIME_IMAGE_REGISTRY}/${RUNTIME_IMAGE_REPO}:${RUNTIME_IMAGE_TAG}"}
## end parameters

echo "building image \`${run_image_uri}\` with Functions runtime and go worker"
worker_root="${__dirname}/../"
docker build -t "${run_image_uri}" -f "$worker_root"/Dockerfile.bundle "$worker_root"

if [[ ( $publish == 1 ) && ( "$RUNTIME_IMAGE_REGISTRY" != "local" ) ]]; then
    echo "pushing image to registry defined in environment"
    docker push "${run_image_uri}"
elif [[ ( $publish == 1 ) && ( "$RUNTIME_IMAGE_REGISTRY" == "local" ) ]]; then
    echo "not trying to publish because \`local\` registry was specified"
fi

