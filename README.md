# Tailscale JetKVM Plugin
This is a plugin for the JetKVM to add Tailscale support. It implements a simple TLS proxy to expose the JetKVM interface to the tailnet as well as a WebRTC TURN server to provide a mechanism to proxy WebRTC traffic through the tailnet without kernel-level TUN device support.

## Building
Run `GOOS=linux GOARCH=arm GOARM=7 go build .` to build the `jetkvm-plugin-tailscale` binary.

Run `tar -czvf tailscale.tar.gz manifest.json jetkvm-plugin-tailscale` to build the plugin archive