#!/usr/bin/env bash

# prolog
__filename=${BASH_SOURCE[0]}
__dirname=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# end prolog

account_name=$1
group_name=$2
location=$3

container_names=(demo demo-out)
queue_name=demoqueue

group_id=$(ensure_group $group_name)

name_available=$(az storage account check-name \
    -n $account_name \
    --query nameAvailable --output tsv)
debug "storage name $account_name available? $name_available"

debug "creating storage account $account_name"
account_id=$(az storage account show \
    --name $account_name \
    --resource-group $group_name \
    --query id --output tsv)
if [[ -z $account_id ]]; then
    account_id=$(az storage account create \
        --name $account_name \
        --resource-group $group_name \
        --location $location \
        --sku 'Standard_LRS' \
        --query id --output tsv)
fi
debug "ensured storage account: $account_id"

debug "getting account key"
key=$(az storage account keys list \
    --account-name $account_name \
    --resource-group $group_name \
    --query '[0].value' -o tsv)

debug "getting account connstr"
connstr=$(az storage account show-connection-string \
            --ids $account_id \
            --query connectionString --output tsv)

debug "creating containers"
for container_name in ${container_names[@]}; do
    az storage container create \
        --name $container_name \
        --account-key $key \
        --account-name $account_name \
        --output tsv --query name
done

debug "creating queues"
az storage queue create \
    --name $queue_name \
    --account-key $key \
    --account-name $account_name \
    --query name --output tsv

echo "$connstr"

