// internal/service/notification_service.go
package service

import (
	"context"
	"log"

	"github.com/jaygaha/medication-tracker-api/internal/models"
	"github.com/jaygaha/medication-tracker-api/internal/notification"
	"github.com/jaygaha/medication-tracker-api/internal/repository"
)

type NotificationService struct {
	apns      notification.Provider
	fcm       notification.Provider
	tokenRepo repository.DeviceTokenRepository
}

func NewNotificationService(apns, fcm notification.Provider, tokenRepo repository.DeviceTokenRepository) *NotificationService {
	return &NotificationService{
		apns:      apns,
		fcm:       fcm,
		tokenRepo: tokenRepo,
	}
}

func (s *NotificationService) SendDoseReminder(ctx context.Context, userID, medicationName, timeOfDay string) error {
	tokens, err := s.tokenRepo.GetActiveTokensByUserID(ctx, userID)
	if err != nil {
		log.Printf("[NotificationService] Error getting tokens for user %s: %v", userID, err)
		return err
	}

	if len(tokens) == 0 {
		log.Printf("[NotificationService] No active tokens for user %s. Skipping reminder for %s.", userID, medicationName)
		return nil
	}

	for _, t := range tokens {
		msg := notification.Message{
			Token: t.Token,
			Title: "Medication Reminder",
			Body:  "It's time to take your " + medicationName + " (" + timeOfDay + ")",
			Data:  map[string]string{"medication": medicationName, "time": timeOfDay},
		}

		var sendErr error
		if t.Platform == models.DevicePlatformIOS {
			sendErr = s.apns.Send(ctx, msg)
		} else {
			sendErr = s.fcm.Send(ctx, msg)
		}

		if sendErr != nil {
			log.Printf("[NotificationService] Error sending to token %s: %v", t.Token, sendErr)
			// Depending on error, we might deactivate token here
		}
	}
	return nil
}
