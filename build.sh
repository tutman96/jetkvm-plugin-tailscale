#!/bin/bash

set -e

VERSION=$(jq -r '.version' manifest.json)
GOOS=linux GOARCH=arm GOARM=7 go build --ldflags="-X main.version=$VERSION" -o jetkvm-plugin-tailscale .
tar -czvf tailscale.tar.gz manifest.json jetkvm-plugin-tailscale