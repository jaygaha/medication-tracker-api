// internal/notification/apns.go
package notification

import (
	"context"
	"log"

	"github.com/jaygaha/medication-tracker-api/internal/config"
)

type APNsProvider struct {
	cfg *config.Config
}

func NewAPNsProvider(cfg *config.Config) Provider {
	return &APNsProvider{cfg: cfg}
}

func (p *APNsProvider) Send(ctx context.Context, msg Message) error {
	log.Printf("[APNsProvider] Sending notification to iOS device token %s: %s - %s", msg.Token, msg.Title, msg.Body)
	// In a real implementation, we would use github.com/sideshow/apns2 here
	// and authenticate with a .p8 key. For now, it's a no-op that logs.
	return nil
}

func (p *APNsProvider) Name() string {
	return "APNs"
}
