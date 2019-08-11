package aggregator

import "time"

const (
	LOW = iota
	MEDIUM
	HIGH
)

type Aggregation []*SecurityNotification

func Strategy(evt *SecurityNotification, s Aggregation) (*AggregateNotification, Aggregation) {
	s = append(s, evt)
	if len(s) == 3 || evt.Priority == HIGH {
		return &AggregateNotification{
			Email:         evt.Email,
			Notifications: s,
		}, nil
	}
	return nil, s
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
