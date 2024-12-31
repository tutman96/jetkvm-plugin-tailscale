package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/caarlos0/env"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/tutman96/jetkvm-plugin-tailscale/plugin"
	"tailscale.com/client/tailscale"
	"tailscale.com/tsnet"
)

type PluginImpl struct {
	client          *jsonrpc2.Conn
	tailscaleServer *tsnet.Server
	tailscaleClient *tailscale.LocalClient
}

var Config struct {
	PluginSocket     string `env:"JETKVM_PLUGIN_SOCK" envDefault:"./tmp/plugin.sock"`
	PluginWorkingDir string `env:"JETKVM_PLUGIN_WORKING_DIR" envDefault:"./tmp"`
}

func connect(ctx context.Context) (*PluginImpl, error) {
	conn, err := net.Dial("unix", Config.PluginSocket)
	if err != nil {
		return nil, err
	}

	impl := &PluginImpl{}
	client := jsonrpc2.NewConn(ctx, jsonrpc2.NewBufferedStream(conn, jsonrpc2.VSCodeObjectCodec{}), plugin.HandleRPC(impl))
	impl.client = client

	return impl, nil
}

func main() {
	env.Parse(&Config)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)
		<-signalChan
		fmt.Println("Received an interrupt, stopping services...")
		cancel()
	}()

	impl, err := connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer impl.client.Close()

	log.Println("client started")

	err = impl.NewTsServer()
	if err != nil {
		log.Fatal(err)
	}
	defer impl.tailscaleServer.Close()

	turnServer, err := impl.CreateTurnServer(ctx)
	if err != nil {
		log.Fatal(err) // TODO: gracefully handle error now that the server can report it
	}
	defer turnServer.Close()

	<-ctx.Done()
}