package webhook

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

const (
	//tlsDir      = `/run/secrets/tls`
	tlsDir      = `/tmp/k8s-webhook-server/serving-certs`
	tlsCertFile = `tls.crt`
	tlsKeyFile  = `tls.key`
)

// RunWebhookServer represents initialization of an http server
func RunWebhookServer() error {
	certPath := filepath.Join(tlsDir, tlsCertFile)
	keyPath := filepath.Join(tlsDir, tlsKeyFile)
	server := newWebhookServer()

	go func() {
		log.Print("Starting webhook server...")
		log.Fatal(server.ListenAndServeTLS(certPath, keyPath))
	}()

	// listening shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Print("Got shutdown signal, shutting down webhook server gracefully...")
	return server.Shutdown(context.Background())
}

func newWebhookServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", quarantineHandler)
	return &http.Server{
		// We listen on port 9443 such that we do not need root privileges or extra capabilities for this server.
		// The Service object will take care of mapping this port to the HTTPS port 443.
		Addr:    ":9443",
		Handler: mux,
	}
}
