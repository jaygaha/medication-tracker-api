// internal/notification/provider.go
// the abstraction that both APNs and FCM implement
package notification

import "context"

// Message struct represents a push notification payload
type Message struct {
	Token string            // Device Token
	Title string            // Notification title (shown to user)
	Body  string            // Notification body (shown to user)
	Data  map[string]string // Custom payload (e.g., medication_id)
}

type Provider interface {
	Send(ctx context.Context, msg Message) error
	Name() string
}
