#!/usr/bin/env bash
set -e
set -o pipefail

rm -rf ./output
go run . g -i
# scp -r ./output lyyyuna:/home/ubuntu/output
rsync -avz --delete ./output lyyyuna:/home/ubuntu