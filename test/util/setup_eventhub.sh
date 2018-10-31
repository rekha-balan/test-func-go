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

debug "ensuring event hub namespace $namespace_name"
namespace_id=$(az eventhubs namespace show \
    --name $namespace_name \
    --resource-group $group_name \
    --query id --output tsv)
if [[ -z $namespace_id ]]; then
    namespace_id=$(az eventhubs namespace create \
        --name $namespace_name \
        --resource-group $group_name \
        --location $location \
        --sku 'Standard' \
        --query id --output tsv)
fi
debug "ensured event hub namespace: $namespace_id"

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

debug "ensuring hubs"
for eventhub_name in ${eventhub_names[@]}; do
    hub_name=$(az eventhubs eventhub show \
        --name $eventhub_name \
        --namespace-name $namespace_name \
        --resource-group $group_name \
        --output tsv --query name)
    if [ -z $hub_name ]; then
        hub_name=$(az eventhubs eventhub create \
            --name $eventhub_name \
            --namespace-name $namespace_name \
            --resource-group $group_name \
            --output tsv --query name)
    fi
    debug "ensured hub: $hub_name"
done

echo $connstr
