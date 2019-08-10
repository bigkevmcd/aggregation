package main

import "time"

const (
	LOW = iota
	MEDIUM
	HIGH
)

type Aggregation []*SecurityNotification

func Strategy(evt *SecurityNotification, s Aggregation) (*AggregateNotification, Aggregation) {
	if s == nil {
		s = make(Aggregation, 0)
	}
	s = append(s, evt)
	if len(s) == 3 || evt.Priority == HIGH {
		return &AggregateNotification{
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
	Notifications []*SecurityNotification
}
