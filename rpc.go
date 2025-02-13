package main

import (
	"context"

	"github.com/TheTechNetwork/jetkvm-plugin-tailscale/plugin"
)

func (p *PluginImpl) GetPluginSupportedMethods(ctx context.Context) (plugin.SupportedMethodsResponse, error) {
	return plugin.SupportedMethodsResponse{
		SupportedRpcMethods: []string{"getPluginSupportedMethods", "getPluginStatus"},
	}, nil
}

func (p *PluginImpl) GetPluginStatus(ctx context.Context) (plugin.PluginStatus, error) {
	if p.tailscaleClient == nil {
		return plugin.PluginStatus{
			Status: "loading",
		}, nil
	}

	status, err := p.tailscaleClient.Status(ctx)
	if err != nil {
		errStr := err.Error()
		return plugin.PluginStatus{
			Status:  "error",
			Message: &errStr,
		}, err
	}

	if status.BackendState == "Running" {
		return plugin.PluginStatus{
			Status: "running",
		}, nil
	} else if status.BackendState == "NeedsLogin" {
		message := "Finish setting up the device at " + status.AuthURL
		return plugin.PluginStatus{
			Status:  "pending-configuration",
			Message: &message,
		}, nil
	}

	return plugin.PluginStatus{
		Status: "loading",
	}, nil
}
