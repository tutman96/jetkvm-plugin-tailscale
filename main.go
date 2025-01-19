package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"

	// This is needed since Tailscale issues LetsEncrypt certificates and validates them against
	// the golang system store, which has no ability to override in the Go runtime (for obvious reasons).
	// See https://github.com/tailscale/tailscale/issues/14690
	_ "github.com/breml/rootcerts"
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

var version = "dev"

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
	client := jsonrpc2.NewConn(ctx, jsonrpc2.NewPlainObjectStream(conn), plugin.HandleRPC(impl))
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

	log.Default().SetPrefix("[jetkvm-plugin-tailscale v" + version + "] ")

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

	cmd := exec.CommandContext(ctx, "/sbin/ifconfig", "lo", "127.0.0.1", "up")
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	err = impl.CreateProxyServer(ctx)
	if err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
}
