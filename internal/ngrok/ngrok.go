package ngrok

import (
	"context"
	"fmt"
	"os"

	"golang.ngrok.com/ngrok/v2"
)

// RunNgrok starts an ngrok tunnel forwarding to the specified local address, returns the public URL, or an error if it fails.
func RunNgrok(ctx context.Context, address string, urlChan chan<- string, errChan chan<- error) {
	ngrokAuthToken := os.Getenv("NGROK_AUTHTOKEN")
	if ngrokAuthToken == "" {
		errChan <- fmt.Errorf("NGROK_AUTHTOKEN environment variable is not set, visit https://dashboard.ngrok.com/get-started/your-authtoken to obtain one")
		return
	}
	agent, err := ngrok.NewAgent(ngrok.WithAuthtoken(ngrokAuthToken))
	if err != nil {
		errChan <- fmt.Errorf("failed to create ngrok agent: %w", err)
		return
	}

	ln, err := agent.Forward(ctx,
		ngrok.WithUpstream(address),
		ngrok.WithURL(os.Getenv("NGROK_RESERVED_DOMAIN")),
	)

	if err != nil {
		errChan <- fmt.Errorf("failed to start ngrok tunnel: %w", err)
		return
	}

	// fmt.Println("Endpoint online: forwarding from", ln.URL(), "to", address)
	urlChan <- ln.URL().String()

	// Explicitly stop forwarding; otherwise it runs indefinitely
	<-ln.Done()
}