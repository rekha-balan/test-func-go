#!/usr/bin/env bash

## prolog
set -o errexit
__filename=${BASH_SOURCE[0]}
__dirname=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
__root=$(cd "${__dirname}/../" && pwd)
if [[ ! -f "${__root}/.env" ]]; then cp "${__root}/.env.tpl" "${__root}/.env"; fi
source "${__root}/.env"
source "${__dirname}/util/helpers.sh"
export -f debug # from helpers.sh, echo to /dev/stderr
## end prolog

function_dir_path=$1
instance_name=${2:-${RUNTIME_INSTANCE_NAME}}
group_name=${3:-"${AZURE_GROUP_NAME_BASE}-smoker"}
user_name=${4:-"function-deployer"}
GH_USER=${5:-${GH_USER}}
GH_TOKEN=${6:-${GH_TOKEN}}
password=a${RANDOM}${RANDOM}${RANDOM}Z
function_dir=$(basename $function_dir_path)
function_dir_path=$(cd $function_dir_path && pwd)

functionapp_id=$(az functionapp show \
    --name $instance_name --resource-group $group_name --query 'id' --output tsv)

## building without container leads to error loading plugin:
##    "plugin was built with a different version of package errors"
#
# pushd ${function_dir_path}
# mkdir -p ./bin
# GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -buildmode=plugin \
#     -o "bin/${function_dir}.so" "main.go"
# popd

## build plugin in container which matches worker
docker run -it --rm \
    -v ${function_dir_path}:/go/src/${function_dir} \
    -w /go/src/${function_dir} \
    golang:1.10 \
    /bin/bash -c " \
    mkdir -p /go/src/github.com/vladbarosan/test-func-go; \
    git clone https://${GH_USER}:${GH_TOKEN}@github.com/vladbarosan/test-func-go \
        /go/src/github.com/vladbarosan/test-func-go;
    go build -buildmode=plugin -o bin/${function_dir}.so main.go"

az functionapp deployment user set \
    --user-name $user_name \
    --password $password > /dev/null

# returns URL with /site/wwwroot appended
ftp_url=$(az functionapp show \
    --name ${instance_name} \
    --resource-group $group_name \
    --query 'ftpPublishingUrl' --output tsv)
ftp_user="${instance_name}\\${user_name}"
ftp_password=$password

# must be set to false for upload of .so to succeed
az functionapp config appsettings set \
    --ids $functionapp_id \
    --settings "WEBSITES_ENABLE_APP_SERVICE_STORAGE=false"
debug "sleeping for 120s"
sleep 120 # wait for setting to be applied...

debug "uploading to ${ftp_url}/${function_dir}/"
debug "with user [${ftp_user}]; password [$ftp_password]"
curl \
    --ftp-create-dirs \
    --user "${ftp_user}:${ftp_password}" \
    --upload-file ${function_dir_path}/main.go \
    --url ${ftp_url}/${function_dir}/main.go \
    --upload-file ${function_dir_path}/function.json \
    --url ${ftp_url}/${function_dir}/function.json \
    --upload-file ${function_dir_path}/bin/${function_dir}.so \
    --url ${ftp_url}/${function_dir}/bin/${function_dir}.so \
    --upload-file ${__root}/sample/host.json \
    --url ${ftp_url}/host.json

# another option to consider
# az functionapp deployment source config-zip --src ./path/to/zip.zip --ids $functionapp_id

# must be set to true to make user function usable
az functionapp config appsettings set \
    --ids $functionapp_id \
    --settings "WEBSITES_ENABLE_APP_SERVICE_STORAGE=true"
debug "sleeping for 120s..."
sleep 120 # wait for setting to be applied

