// internal/service/scheduler_service.go
package service

import (
	"context"
	"log"
	"time"

	"github.com/jaygaha/medication-tracker-api/internal/models"
	"github.com/jaygaha/medication-tracker-api/internal/repository"
)

type SchedulerService struct {
	scheduleRepo        repository.ScheduleRepository
	medRepo             repository.MedicationRepository
	notifService        *NotificationService
	notificationLogRepo repository.NotificationLogRepository
}

func NewSchedulerService(scheduleRepo repository.ScheduleRepository, medRepo repository.MedicationRepository, notifService *NotificationService, notificationLogRepo repository.NotificationLogRepository) *SchedulerService {
	return &SchedulerService{
		scheduleRepo:        scheduleRepo,
		medRepo:             medRepo,
		notifService:        notifService,
		notificationLogRepo: notificationLogRepo,
	}
}

func (s *SchedulerService) Start(ctx context.Context, interval time.Duration) {
	log.Printf("[Scheduler] Starting scheduler worker with interval %v", interval)
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("[Scheduler] Stopping scheduler worker")
				ticker.Stop()
				return
			case <-ticker.C:
				s.processSchedules(ctx)
			}
		}
	}()
}

func (s *SchedulerService) processSchedules(ctx context.Context) {
	log.Println("[Scheduler] Running schedule check")

	schedules, err := s.scheduleRepo.GetAllActiveSchedules(ctx)
	if err != nil {
		log.Printf("[Scheduler] Error fetching active schedules: %v", err)
		return
	}

	now := time.Now()

	for _, sched := range schedules {
		// Timezone Awareness
		loc, err := time.LoadLocation(sched.UserTimeZone)
		if err != nil || sched.UserTimeZone == "" {
			loc = time.UTC
		}
		
		userNow := now.In(loc)
		currentDay := int(userNow.Weekday()) // 0=Sunday

		// 1. Check if today is a valid day for this schedule
		if !s.isDayValid(sched, userNow, currentDay) {
			continue
		}

		// 2. Check if any time matches
		for _, st := range sched.Times {
			parsedTime, err := time.Parse("15:04", st.TimeOfDay)
			if err != nil {
				continue
			}

			scheduledTimeToday := time.Date(userNow.Year(), userNow.Month(), userNow.Day(), parsedTime.Hour(), parsedTime.Minute(), 0, 0, loc)

			// Calculate diff using actual times
			timeDiff := now.Sub(scheduledTimeToday)
			
			if timeDiff >= 0 && timeDiff < 5*time.Minute {
				// Deduplication check
				scheduledDate := time.Date(userNow.Year(), userNow.Month(), userNow.Day(), 0, 0, 0, 0, time.UTC)
				hasSent, err := s.notificationLogRepo.HasSentNotification(ctx, st.ID, scheduledDate)
				
				if err == nil && !hasSent {
					s.triggerNotification(ctx, sched, st, scheduledDate, now)
				} else if err != nil {
					log.Printf("[Scheduler] Error checking deduplication log: %v", err)
				}
			}
		}
	}
}

func (s *SchedulerService) triggerNotification(ctx context.Context, sched *models.Schedule, st models.ScheduleTime, scheduledDate time.Time, now time.Time) {
	med, err := s.medRepo.GetMedicationByID(ctx, sched.MedicationID)
	if err != nil {
		log.Printf("[Scheduler] Error getting medication %s: %v", sched.MedicationID, err)
		return
	}

	err = s.notifService.SendDoseReminder(ctx, med.UserID, med.Name, st.TimeOfDay)
	if err != nil {
		log.Printf("[Scheduler] Error sending reminder for %s: %v", med.Name, err)
		return
	}

	// Log success to prevent double-sending
	logRecord := &models.NotificationLog{
		UserID:         med.UserID,
		ScheduleID:     sched.ID,
		ScheduleTimeID: st.ID,
		ScheduledDate:  scheduledDate,
		SentAt:         now,
	}
	if err := s.notificationLogRepo.RecordNotification(ctx, logRecord); err != nil {
		log.Printf("[Scheduler] Error recording notification sent log: %v", err)
	}
}

func (s *SchedulerService) isDayValid(sched *models.Schedule, now time.Time, currentDay int) bool {
	switch sched.Type {
	case models.FrequencyEveryDay:
		return true
	case models.FrequencySpecificDays:
		for _, d := range sched.Days {
			if d == currentDay {
				return true
			}
		}
		return false
	case models.FrequencyRegularIntervals:
		if !sched.StartDate.IsZero() {
			// Calculate days since start date
			start := sched.StartDate
			// Normalize to midnight
			startMidnight := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
			nowMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			daysSince := int(nowMidnight.Sub(startMidnight).Hours() / 24)
			
			if daysSince >= 0 && sched.IntervalDays != nil && *sched.IntervalDays > 0 && daysSince%(*sched.IntervalDays) == 0 {
				return true
			}
		}
		return false
	case models.FrequencyAsNeeded:
		return false
	default:
		return false
	}
}
