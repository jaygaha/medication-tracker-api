// internal/notification/fcm.go
package notification

import (
	"context"
	"log"

	"github.com/jaygaha/medication-tracker-api/internal/config"
)

type FCMProvider struct {
	cfg *config.Config
}

func NewFCMProvider(cfg *config.Config) Provider {
	return &FCMProvider{cfg: cfg}
}

func (p *FCMProvider) Send(ctx context.Context, msg Message) error {
	log.Printf("[FCMProvider] Sending notification to Android/Web device token %s: %s - %s", msg.Token, msg.Title, msg.Body)
	// In a real implementation, we would use firebase.google.com/go/v4 here
	// and authenticate with a JSON service account key. For now, it's a no-op that logs.
	return nil
}

func (p *FCMProvider) Name() string {
	return "FCM"
}
