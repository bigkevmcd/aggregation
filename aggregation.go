package aggregator

import (
	"time"
)

const (
	LOW = iota
	MEDIUM
	HIGH
)

var clock = func() time.Time {
	return time.Now().UTC()
}

type Aggregation struct {
	Email         string
	LastSaved     time.Time
	Notifications []*SecurityNotification
}

func Strategy(evt *SecurityNotification, s *Aggregation) (*AggregateNotification, *Aggregation) {
	if evt == nil {
		return processAggregationWithoutEvent(s)
	}
	s.Notifications = append(s.Notifications, evt)
	if len(s.Notifications) == 3 || evt.Priority == HIGH {
		return aggregationToNotification(s), nil
	}
	return nil, s
}

func processAggregationWithoutEvent(s *Aggregation) (*AggregateNotification, *Aggregation) {
	cutOffTime := clock().Add(time.Hour * -3)
	if s.LastSaved.Before(cutOffTime) {
		return aggregationToNotification(s), nil
	}
	return nil, s
}

func aggregationToNotification(s *Aggregation) *AggregateNotification {
	return &AggregateNotification{
		Email:         s.Email,
		Notifications: s.Notifications,
	}
}

type SecurityNotification struct {
	Email        string
	Notification string
	Timestamp    time.Time
	Priority     int
}

type AggregateNotification struct {
	Email         string
	Notifications []*SecurityNotification
}

func TranslatePriority(p int) string {
	switch p {
	case LOW:
		return "low"
	case MEDIUM:
		return "medium"
	case HIGH:
		return "high"
	default:
		return "unknown"
	}
}
