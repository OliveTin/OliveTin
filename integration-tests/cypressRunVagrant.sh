#!/bin/bash

# args:
# $1: The Vagrant VM to test against. If blank and only one VM is provisioned, it will use that.

IP=$(vagrant ssh-config $1 | grep HostName | awk '{print $2}')
BASE_URL="http://$IP:1337/"

echo "IP: $IP, BaseURL: $BASE_URL"

# Only run the general test, as we cannot easily switch out configs in VMs yet.
./node_modules/.bin/cypress run --headless -c baseUrl=$BASE_URL -s cypress/e2e/general/*
