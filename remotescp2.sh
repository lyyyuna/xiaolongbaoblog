#!/usr/bin/env bash
set -e
set -o pipefail

rm -rf ./output
go run . g -i
rsync -avz --delete ./output hao:/home/ubuntu