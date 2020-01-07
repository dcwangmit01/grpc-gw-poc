#!/usr/bin/env bash

# This command looks up the name of the binary using the go.mod package name
CLI=$(grep '^module' $(dirname $(dirname $(readlink -f $0)))/go.mod | awk -F/ '{print $NF}')
