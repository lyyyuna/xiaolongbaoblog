#!/usr/bin/env bash
set -e
set -o pipefail

rm -rf ./output
go run . g -i
go run . d