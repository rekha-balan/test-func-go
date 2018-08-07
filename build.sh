#!/usr/bin/env bash

# mount pwd into container and build there
# add `-gcflags '-N -l'` to 'go build ...' to compile for debugging

mode=$1 # use 'native' to do a native build , 'docker' to build through docker(default).
verbose=$2
bundle=$3 # use 'bundle' to build the samples also, otherwise only the worker will be built.

if [ "$verbose" == 'verbose' ]; then
    set -ev
else
    set -e
fi

if [ "$mode" == 'native' ]; then
    echo "building natively..."
    env GOOS=linux GOARCH=amd64 go build -o workers/golang/golang-worker
else
    echo "building worker..."
    docker run -it \
        -v $(pwd):/go/src/github.com/Azure/azure-functions-go \
        -w /go/src/github.com/Azure/azure-functions-go \
         golang:1.10 /bin/bash -c "go build -o workers/golang/golang-worker"
fi

echo "worker built"

if [ "$bundle" == 'bundle' ]; then
    echo "building samples..."
    for file in sample/*/ ; do
        if [ -f $file/function.json ]; then
            s=$(basename $file)
            echo "building sample $s"
            if [ "$mode" == 'native' ]; then
                env GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -buildmode=plugin -o "sample/$s/bin/$s.so" "sample/$s/main.go"
            else
                docker run -it \
                    -v $(pwd):/go/src/github.com/Azure/azure-functions-go \
                    -w /go/src/github.com/Azure/azure-functions-go \
                     golang:1.10 /bin/bash -c "go build -buildmode=plugin -o sample/$s/bin/$s.so sample/$s/main.go"
            fi
        fi
    done
fi

chmod +rx $(pwd)/workers/golang/golang-worker
