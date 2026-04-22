package authclient

import (
	"context"
	"time"
)

func PublicKeyFetcher(
	ctx context.Context,
	client *Client,
	keysChan chan []string,
) {
	// Keep trying until we have fetched public keys
	// One acquired we enter the second loop which periodically Checks
	// for new keys
	for {
		keys, err := client.PublicKeys()
		if err == nil {
			keysChan <- keys
			break
		}
		time.Sleep(time.Second)
	}

	ticker := time.NewTicker(1 * time.Hour)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			keys, err := client.PublicKeys()
			if err == nil {
				keysChan <- keys
			}
		}
	}
}
