#!/usr/bin/env bash

## prolog
set -o errexit
__filename=${BASH_SOURCE[0]}
__dirname=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
__root="${__dirname}/../"
if [[ ! -f "${__root}/.env" ]]; then cp "${__root}/.env.tpl" "${__root}/.env"; fi
source "${__dirname}/../.env"
## end prolog

## parameters
declare -i publish=${1:-0}  # 0: false; 1: true
declare bundle=${2:-"sample"} # "sample" or "usr"
declare run_image_uri=${3:-"${RUNTIME_IMAGE_REGISTRY}/${RUNTIME_IMAGE_REPO}:${RUNTIME_IMAGE_TAG}"}
## end parameters

echo "building image \`${run_image_uri}\` with Functions runtime and go worker"

docker build \
    --tag "${run_image_uri}" \
    --file "${__root}/Dockerfile.bundle" \
    --build-arg "bundle=${bundle}" \
    "${__root}"

if [[ ( $publish == 1 ) && ( "$RUNTIME_IMAGE_REGISTRY" != "local" ) ]]; then
    echo "pushing image to registry defined in environment"
    docker push "${run_image_uri}"
elif [[ ( $publish == 1 ) && ( "$RUNTIME_IMAGE_REGISTRY" == "local" ) ]]; then
    echo "not trying to publish because \`local\` registry was specified"
fi

