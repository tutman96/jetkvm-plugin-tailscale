package main

// Temporary implementation of a bare bones JetKVM plugin server
import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/sourcegraph/jsonrpc2"
)

type ServerHandler struct{}

func (h *ServerHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	// Handle incoming requests
	if req.Method == "ping" {
		conn.Reply(ctx, req.ID, "pong")
		return
	}
	conn.Reply(ctx, req.ID, fmt.Sprintf("Unknown method: %s", req.Method))
}

func main() {
	socketPath := "./tmp/plugin.sock"

	// Remove old socket if it exists
	os.Remove(socketPath)

	// Start listening on the Unix socket
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to listen on socket: %v", err)
	}
	defer listener.Close()

	log.Println("Server is listening...")

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		// Handle the connection with a server
		stream := jsonrpc2.NewBufferedStream(conn, jsonrpc2.VSCodeObjectCodec{})
		client := jsonrpc2.NewConn(context.Background(), stream, &ServerHandler{})

		go func() {
			ticker := time.NewTicker(500 * time.Millisecond)

			for range ticker.C {
				var results interface{}
				err = client.Call(context.Background(), "getPluginStatus", nil, &results)
				if err == jsonrpc2.ErrClosed {
					log.Println("Connection closed")
					break
				}
				if err != nil {
					log.Printf("Failed to call method: %v", err)
				}
				log.Printf("Result: %v", results)
			}
		}()
	}
}
