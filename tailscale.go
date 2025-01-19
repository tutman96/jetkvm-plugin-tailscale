package main

import (
	"log"

	"tailscale.com/tsnet"
)

func (p *PluginImpl) NewTsServer() error {
	p.tailscaleServer = new(tsnet.Server)
	p.tailscaleServer.Hostname = "jetkvm" // TODO: make this configurable
	p.tailscaleServer.Dir = Config.PluginWorkingDir
	p.tailscaleServer.Logf = log.Printf
	err := p.tailscaleServer.Start()
	if err != nil {
		return err
	}

	p.tailscaleClient, err = p.tailscaleServer.LocalClient()
	if err != nil {
		return err
	}

	return nil
}
