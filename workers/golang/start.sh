#! /bin/bash

# Directory name for start.sh
DIR="$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)"
echo "executing $DIR/golang-worker $@"
$DIR/golang-worker $@
# Uncomment the next line and comment the previous one for debugging
#$HOME/go/bin/dlv --listen=:40000 --headless=true --api-version=2 exec $DIR/golang-worker -- $@
