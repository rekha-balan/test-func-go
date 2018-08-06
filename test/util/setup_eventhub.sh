#!/usr/bin/env bash

# prolog
__filename=${BASH_SOURCE[0]}
__dirname=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# end prolog

namespace_name=$1
group_name=$2
location=$3

eventhub_names=(demo-go-func-in demo-go-func-out demo-go-func-batch-in)

group_id=$(ensure_group $group_name)
debug "group_id: $group_id"

name_available=$(az eventhubs namespace exists \
    --name $namespace_name \
    --query nameAvailable --output tsv)
debug "event hub namespace name $namespace_name available? $name_available"

debug "creating event hub namespace $namespace_name"
namespace_id=$(az eventhubs namespace create \
    --name $namespace_name \
    --resource-group $group_name \
    --location $location \
    --sku 'Standard' \
    --query id --output tsv)
debug "created event hub namespace: $namespace_id"

debug "getting namespace default SAS policy"
policy_name=$(az eventhubs namespace authorization-rule list \
    --namespace-name $namespace_name \
    --resource-group $group_name \
    --query "[0].name" -o tsv)

debug "getting connstr for SAS policy $policy_name"
connstr=$(az eventhubs namespace authorization-rule keys list \
    --namespace-name $namespace_name \
    --resource-group $group_name \
    --name $policy_name \
    --query "primaryConnectionString" -o tsv)

debug "creating hubs"
for eventhub_name in ${eventhub_names[@]}; do
    hub_name=$(az eventhubs eventhub create \
        --name $eventhub_name \
        --namespace-name $namespace_name \
        --resource-group $group_name \
        --output tsv --query name)
    debug "created hub: $hub_name"
done

echo $connstr
