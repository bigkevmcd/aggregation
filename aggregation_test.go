package main

import (
	"testing"
	"time"
)

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
	if l := len(n.Notifications); l != 3 {
		t.Fatalf("expected 3 messages in the aggregation, got %d", l)
	}
}

func TestAggregationPublishesOnHighPriorityEvent(t *testing.T) {
	n, s := Strategy(makeNotification("a@example.com"), make(state, 0))
	if n != nil {
		t.Fatalf("unexpectedly received a notification: got %#v", n)
	}
	evt2 := makeNotification("a@example.com")
	evt2.Priority = HIGH
	n, s = Strategy(evt2, s)
	if n == nil {
		t.Fatal("expected a notification, got nil")
	}

	if l := len(n.Notifications); l != 2 {
		t.Fatalf("expected 2 messages in the aggregation, got %d", l)
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
