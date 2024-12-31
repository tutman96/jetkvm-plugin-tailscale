package plugin

import (
	"context"
	"log"

	"github.com/sourcegraph/jsonrpc2"
)

type PluginStatus struct {
	Status string  `json:"status" oneOf:"running,loading,pending-configuration,error"`
	Error  *string `json:"error,omitempty"`
}

type PluginHandler interface {
	GetPluginSupportedMethods(ctx context.Context) ([]string, error)
	GetPluginStatus(ctx context.Context) (PluginStatus, error)
}

func HandleRPC(handler PluginHandler) jsonrpc2.Handler {
	return jsonrpc2.HandlerWithError(func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
		log.Printf("Received request: %s", req.Method)
		switch req.Method {
		case "getPluginSupportedMethods":
			return handler.GetPluginSupportedMethods(ctx)
		case "getPluginStatus":
			return handler.GetPluginStatus(ctx)
		}
		return nil, nil
	})
}
