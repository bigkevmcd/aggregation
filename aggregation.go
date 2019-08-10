package main

import "time"

const (
	LOW    = iota
	MEDIUM = iota
	HIGH   = iota
)

type state []*SecurityNotification

func Strategy(evt *SecurityNotification, s state) (*AggregateNotification, state) {
	s = append(s, evt)
	if len(s) == 3 {
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
