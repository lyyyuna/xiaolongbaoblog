#!/usr/bin/env bash
set -e
set -o pipefail

ssh lyyyuna "rm -rf /home/ubuntu/blog"
rm -rf ./output
scp -r ./output lyyyuna:/home/ubuntu/blog