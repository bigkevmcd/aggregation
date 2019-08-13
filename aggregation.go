package aggregator

import "time"

const (
	LOW = iota
	MEDIUM
	HIGH
)

var clock = func() time.Time {
	return time.Now().UTC()
}

type Aggregation []*SecurityNotification

func Strategy(evt *SecurityNotification, s Aggregation) (*AggregateNotification, Aggregation) {
	if evt == nil {
		return processAggregationWithoutEvent(s)
	}
	s = append(s, evt)
	if len(s) == 3 || evt.Priority == HIGH {
		return &AggregateNotification{
			Email:         evt.Email,
			Notifications: s,
		}, nil
	}
	return nil, s
}

func processAggregationWithoutEvent(s Aggregation) (*AggregateNotification, Aggregation) {
	if len(s) == 0 {
		return nil, nil
	}
	cutOffTime := clock().Add(time.Hour * -3)
	earliest := earliestNotification(s)
	if earliest.Before(cutOffTime) {
		return &AggregateNotification{
			Email:         s[0].Email,
			Notifications: s,
		}, nil
	}
	return nil, s
}

func earliestNotification(s Aggregation) time.Time {
	earliest := s[0].Timestamp
	for _, n := range s[1:] {
		if n.Timestamp.Before(earliest) {
			earliest = n.Timestamp
		}
	}
	return earliest
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
