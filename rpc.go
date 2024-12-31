package main

import (
	"context"

	"github.com/tutman96/jetkvm-plugin-tailscale/plugin"
)

func (p *PluginImpl) GetPluginSupportedMethods(ctx context.Context) ([]string, error) {
	return []string{"getPluginSupportedMethods", "getPluginStatus"}, nil
}

func (p *PluginImpl) GetPluginStatus(ctx context.Context) (plugin.PluginStatus, error) {
	status, err := p.tailscaleClient.Status(ctx)
	if err != nil {
		errStr := err.Error()
		return plugin.PluginStatus{
			Status: "error",
			Error:  &errStr,
		}, err
	}

	if status.BackendState == "Running" {
		return plugin.PluginStatus{
			Status: "running",
		}, nil
	} else if status.BackendState == "NeedsLogin" {
		return plugin.PluginStatus{
			Status: "pending-configuration",
		}, nil
	}

	return plugin.PluginStatus{
		Status: "loading",
	}, nil
}
