package main

import "time"

const (
	LOW = iota
	MEDIUM
	HIGH
)

type state []*SecurityNotification

func Strategy(evt *SecurityNotification, s state) (*AggregateNotification, state) {
	if s == nil {
		s = make([]*SecurityNotification, 0)
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
