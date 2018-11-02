#!/usr/bin/env bash

# prolog
__filename=${BASH_SOURCE[0]}
__dirname=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# end prolog

account_name=$1
group_name=$2
location=$3

db_name="Documents"
collection_names=(reports tasks leases)

group_id=$(ensure_group $group_name)
debug "group_id: $group_id"

name_available=$(az cosmosdb check-name-exists \
    --name $account_name \
    --output tsv)
debug "cosmos db account name $account_name available? $name_available"

debug "ensuring cosmosdb account $account_name"
account_id=$(az cosmosdb show \
    --name $account_name \
    --resource-group $group_name \
    --query id --output tsv)
if [[ -z $account_id ]]; then
    account_id=$(az cosmosdb create \
        --name $account_name \
        --resource-group $group_name \
        --kind 'GlobalDocumentDB' \
        --query id --output tsv)
fi
debug "ensured cosmosdb account: $account_id"

debug "getting account key and making connstr"
account_key=$(az cosmosdb list-keys \
    --name $account_name \
    --resource-group $group_name \
    --query "primaryMasterKey" -o tsv)

connstr="AccountEndpoint=https://${account_name}.documents.azure.com:443/;AccountKey=${account_key};"

debug "ensuring cosmosdb database [$db_name]"
db_id=$(az cosmosdb database show \
    --db-name $db_name \
    --name $account_name \
    --resource-group-name $group_name \
    --output tsv --query 'id' 2> /dev/null)
if [[ -z $db_id ]]; then
    db_id=$(az cosmosdb database create \
        --db-name $db_name \
        --name $account_name \
        --resource-group $group_name \
        --output tsv --query 'id')
fi
debug "ensured cosmosdb database: [$db_id]"

debug "ensuring collections"
for collection_name in ${collection_names[@]}; do
    coll_id=$(az cosmosdb collection show \
        --collection-name $collection_name \
        --db-name $db_name \
        --name $account_name \
        --resource-group-name $group_name \
        --output tsv --query 'collection.id' 2> /dev/null)
    if [[ -z $coll_id ]]; then
        coll_id=$(az cosmosdb collection create \
            --collection-name $collection_name \
            --db-name $db_name \
            --name $account_name \
            --resource-group-name $group_name \
            --output tsv --query 'collection.id')
    fi
    debug "ensured collection: [$coll_id]"
done

echo $connstr
