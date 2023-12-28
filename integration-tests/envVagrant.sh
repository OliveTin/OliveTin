#!/bin/bash
# Run this like `. envVagrant.sh f38` before `mocha`

# args:
# $1: The Vagrant VM to test against. If blank and only one VM is provisioned, it will use that.

export IP=$(vagrant ssh-config $1 | grep HostName | awk '{print $2}')
export PORT=1337
