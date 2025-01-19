package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const jetkvmHost = "127.0.0.1:80"

func (p *PluginImpl) CreateProxyServer(ctx context.Context) error {
	// Start :80 plaintext listener
	plaintextListener, err := p.tailscaleServer.Listen("tcp", ":80")
	if err != nil {
		return err
	}

	// Fallback proxy handler for non-TLS requests
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   jetkvmHost,
	})

	useTLS := false
	httpServer := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Redirect to HTTPS if TLS is available, otherwise proxy the request
			if useTLS && r.TLS == nil {
				httpsURL := fmt.Sprintf("https://%s%s", r.Host, r.RequestURI)
				http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
			} else {
				proxy.ServeHTTP(w, r)
			}
		}),
	}

	go serveUntilCancel(ctx, httpServer, plaintextListener)

	// Attempt to start :443 HTTPS listener
	httpsListener, err := p.tailscaleServer.ListenTLS("tcp", ":443")
	if err != nil {
		log.Printf("TLS listener failed: %v. Falling back to plaintext only.", err)
		return nil
	}
	useTLS = true
	go serveUntilCancel(ctx, httpServer, httpsListener)
	return nil
}

func serveUntilCancel(ctx context.Context, srv *http.Server, l net.Listener) error {
	go func() {
		err := srv.Serve(l)
		if err != nil && err != http.ErrServerClosed {
			log.Printf("Error starting server: %v", err)
		}
	}()

	<-ctx.Done()
	return srv.Shutdown(context.Background())
}
