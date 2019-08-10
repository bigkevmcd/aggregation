package main

import (
	"testing"
	"time"
)

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

func TestAggregation(t *testing.T) {
	n, s := Strategy(makeNotification("a@example.com"), make(state, 0))
	if n != nil {
		t.Fatalf("unexpectedly received a notification: got %#v", n)
	}
	n, s = Strategy(makeNotification("a@example.com"), s)
	if n != nil {
		t.Fatalf("unexpectedly received a notification: got %#v", n)
	}
	n, s = Strategy(makeNotification("a@example.com"), s)
	if n == nil {
		t.Fatal("expected a notification, got nil")
	}

}

func makeNotification(email string) *SecurityNotification {
	return &SecurityNotification{
		Email:        email,
		Notification: "testing",
		Timestamp:    time.Now().UTC(),
		Priority:     LOW,
	}
}
