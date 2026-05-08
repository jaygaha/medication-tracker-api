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
	scheduleRepo repository.ScheduleRepository
	medRepo      repository.MedicationRepository
	notifService *NotificationService
}

func NewSchedulerService(scheduleRepo repository.ScheduleRepository, medRepo repository.MedicationRepository, notifService *NotificationService) *SchedulerService {
	return &SchedulerService{
		scheduleRepo: scheduleRepo,
		medRepo:      medRepo,
		notifService: notifService,
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
	// Current day 0=Sunday
	currentDay := int(now.Weekday())

	for _, sched := range schedules {
		// 1. Check if today is a valid day for this schedule
		if !s.isDayValid(sched, now, currentDay) {
			continue
		}

		// 2. Check if any time matches
		for _, st := range sched.Times {
			parsedTime, err := time.Parse("15:04", st.TimeOfDay)
			if err != nil {
				continue
			}

			scheduledTimeToday := time.Date(now.Year(), now.Month(), now.Day(), parsedTime.Hour(), parsedTime.Minute(), 0, 0, now.Location())

			timeDiff := now.Sub(scheduledTimeToday)
			// Assuming interval is 5 mins, we trigger if timeDiff is between 0 and 5 minutes
			if timeDiff >= 0 && timeDiff < 5*time.Minute {
				s.triggerNotification(ctx, sched, st.TimeOfDay)
			}
		}
	}
}

func (s *SchedulerService) triggerNotification(ctx context.Context, sched *models.Schedule, timeOfDay string) {
	med, err := s.medRepo.GetMedicationByID(ctx, sched.MedicationID)
	if err != nil {
		log.Printf("[Scheduler] Error getting medication %s: %v", sched.MedicationID, err)
		return
	}

	err = s.notifService.SendDoseReminder(ctx, med.UserID, med.Name, timeOfDay)
	if err != nil {
		log.Printf("[Scheduler] Error sending reminder for %s: %v", med.Name, err)
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
