#!/bin/bash

set -o xtrace

echo "Running config $1"

cp -f ../configs/config.{$1}.yaml ./config/config.yaml
docker start olivetin
NO_COLOR=1 ./node_modules/.bin/cypress run --headless -s cypress/integration/$1/*  || true
docker kill olivetin

