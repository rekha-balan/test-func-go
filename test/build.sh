#!/usr/bin/env bash

## prolog
set -o errexit
__filename=${BASH_SOURCE[0]}
__dirname=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
__root=$(cd "${__dirname}/../" && pwd)
if [[ ! -f "${__root}/.env" ]]; then cp "${__root}/.env.tpl" "${__root}/.env"; fi
source "${__dirname}/../.env"
## end prolog

mode=${1:-"docker"} # 'native' to build without container, 'docker' to build in container (default)
bundle=${2:-"sample"} # 'sample' to build sample/ (default), 'user' to build user/, empty to only build worker
verbose=$3

if [ "$verbose" == 'verbose' ]; then
    set -ev
else
    set -e
fi

# get deps
DEP_RELEASE_TAG=v0.5.0
curl -sSL https://raw.githubusercontent.com/golang/dep/master/install.sh | sh \
    && dep ensure -v -vendor-only

# build worker
# add `-gcflags '-N -l'` to 'go build ...' to compile for debugging
if [ "$mode" == 'native' ]; then
    echo "building natively..."
    env GOOS=linux GOARCH=amd64 go build -o ${__root}/workers/golang/golang-worker
    chmod +rx $(pwd)/workers/golang/golang-worker
else
    echo "building worker..."
    # mount pwd into container and build there
    docker run -it --rm \
        -v $(pwd):/go/src/github.com/vladbarosan/test-func-go \
        -w /go/src/github.com/vladbarosan/test-func-go \
        golang:1.10 \
        /bin/bash -c "go build -o workers/golang/golang-worker && \
                      chmod +rx workers/golang/golang-worker"
fi
echo "worker built"

# build_function takes a path to function files `function_path`
# where it expects to find `function.json` and `main.go` files comprising
# a user function
function build_function () {
    local function_path=$1
    function_path=$(cd $function_path && pwd) # full path

    if [ -f ${function_path}/function.json ]; then
        function_name=$(basename ${function_path})
        echo "building user function $function_name"
        if [ "$mode" == 'native' ]; then
            env GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
                go build -buildmode=plugin \
                -o "${function_path}/bin/${function_name}.so" \
                "${function_path}/main.go"
        else
            docker run -it --rm \
                -v "${__root}":/go/src/github.com/vladbarosan/test-func-go \
                -v "${function_path}":/go/src/${function_name} \
                -w /go/src/${function_name} \
                golang:1.10 \
                /bin/bash -c "go build -buildmode=plugin \
                    -o bin/${function_name}.so main.go"
        fi
    fi

}

if [ "$bundle" == 'sample' ]; then
    echo "building samples in ${_root}/sample..."
    for file in ${__root}/sample/*/ ; do
        build_function $file
    done
elif [ "$bundle" == 'usr' ]; then
    echo "building user code in ${__root}/usr..."
    for file in ${__root}/usr/*/ ; do
        build_function $file
    done
fi

