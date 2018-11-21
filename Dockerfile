# --build-arg RUNTIME_IMAGE; functions runtime image
ARG RUNTIME_IMAGE=mcr.microsoft.com/azure-functions/base:2.0

# build worker then copy to runtime image
FROM golang:1.11 as golang-env
WORKDIR /go/src/github.com/Azure/azure-functions-go
ENV DEP_RELEASE_TAG=v0.5.0
COPY . .
RUN curl -sSL https://raw.githubusercontent.com/golang/dep/master/install.sh | sh \
    && dep ensure -v -vendor-only \
    && chmod +x ./test/build.sh \
    && ./test/build.sh native none verbose

# 3. copy built worker and extensions to runtime image
# ARG instructions used here must be declared before first FROM
FROM ${RUNTIME_IMAGE}

# copy worker to predefined path
COPY --from=golang-env \
    /go/src/github.com/Azure/azure-functions-go/workers/golang \
    /azure-functions-host/workers/golang/

# use predefined env var names to point to worker start script
ENV workers:golang:path /azure-functions-host/workers/golang/start.sh
# remove the following line to disable console logging. To 
# configure logging, see the host.json file.
ENV AzureFunctionsJobHost__Logging__Console__IsEnabled=true