package plugin

import (
	"context"

	"github.com/sourcegraph/jsonrpc2"
)

type PluginStatus struct {
	Status  string  `json:"status" oneOf:"running,loading,pending-configuration,error"`
	Message *string `json:"message,omitempty"`
}

type SupportedMethodsResponse struct {
	SupportedRpcMethods []string `json:"supported_rpc_methods"`
}

type PluginHandler interface {
	GetPluginSupportedMethods(ctx context.Context) (SupportedMethodsResponse, error)
	GetPluginStatus(ctx context.Context) (PluginStatus, error)
}

func HandleRPC(handler PluginHandler) jsonrpc2.Handler {
	return jsonrpc2.HandlerWithError(func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
		// log.Printf("Received request: %s", req.Method)
		switch req.Method {
		case "getPluginSupportedMethods":
			return handler.GetPluginSupportedMethods(ctx)
		case "getPluginStatus":
			return handler.GetPluginStatus(ctx)
		}
		return nil, nil
	})
}
