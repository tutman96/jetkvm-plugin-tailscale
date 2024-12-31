package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/pion/turn/v4"
)

func (p *PluginImpl) waitForIP(ctx context.Context) (net.IP, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled")
		case <-ticker.C:
			log.Printf("Waiting for Tailscale to obtain an ip...")
			ipv4Addr, _ := p.tailscaleServer.TailscaleIPs()
			slice := ipv4Addr.AsSlice()
			if slice == nil {
				continue
			}
			return net.IP(slice), nil
		}
	}
}

func (p *PluginImpl) CreateTurnServer(ctx context.Context) (*turn.Server, error) {
	// Wait for a valid ip address
	ipv4, err := p.waitForIP(ctx)
	if err != nil {
		return nil, err
	}
	log.Printf("Tailscale IPs: %v", ipv4)

	listenerAddress := fmt.Sprintf("%s:3478", ipv4.String())
	udpListener, err := p.tailscaleServer.ListenPacket("udp", listenerAddress)
	if err != nil {
		return nil, err
	}

	// TODO: this could be dynamic depending on the JetKVM side of things
	key := turn.GenerateAuthKey("username", "pion.ly", "password")
	s, err := turn.NewServer(turn.ServerConfig{
		Realm: "pion.ly",
		AuthHandler: func(username string, realm string, srcAddr net.Addr) ([]byte, bool) {
			log.Printf("Authenticating %s", username)
			if username == "username" {
				return key, true
			}
			return nil, false
		},
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: udpListener,
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					RelayAddress: ipv4, // IPv4 only for now
					Address:      "0.0.0.0",
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	log.Printf("TURN server listening on %s", listenerAddress)

	return s, nil
}
