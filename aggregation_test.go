package aggregator

import (
	"testing"
	"time"
)

const testEmail = "a@example.com"

func TestAggregation(t *testing.T) {
	n, s := Strategy(makeNotification(testEmail), makeAggregation())
	if n != nil {
		t.Fatalf("unexpectedly received a notification: got %#v", n)
	}
	n, s = Strategy(makeNotification(testEmail), s)
	if n != nil {
		t.Fatalf("unexpectedly received a notification: got %#v", n)
	}
	n, s = Strategy(makeNotification(testEmail), s)
	if n == nil {
		t.Fatal("expected a notification, got nil")
	}
	if n.Email != testEmail {
		t.Fatalf("incorrect aggregate email: got %s, wanted %s", n.Email, testEmail)
	}
	if l := len(n.Notifications); l != 3 {
		t.Fatalf("expected 3 messages in the aggregation, got %d", l)
	}
}

func TestAggregationPublishesOnHighPriorityEvent(t *testing.T) {
	n, s := Strategy(makeNotification(testEmail), makeAggregation())
	if n != nil {
		t.Fatalf("unexpectedly received a notification: got %#v", n)
	}
	evt2 := makeNotification(testEmail)
	evt2.Priority = HIGH
	n, s = Strategy(evt2, s)
	if n == nil {
		t.Fatal("expected a notification, got nil")
	}

	if l := len(n.Notifications); l != 2 {
		t.Fatalf("expected 2 messages in the aggregation, got %d", l)
	}
}

func TestAggregationWithoutEventAndEmptyState(t *testing.T) {
	n, s := Strategy(nil, makeAggregation())
	if n != nil {
		t.Fatalf("unexpectedly received a notification: got %#v", n)
	}
	if s != nil {
		t.Fatalf("got aggregation state: %#v", s)
	}
}

func TestAggregationWithoutEventOldNotifications(t *testing.T) {
	oldNotification := makeNotification(testEmail)
	oldNotification.Timestamp = time.Now().UTC().Add(time.Hour * -4)

	a := makeAggregation(makeNotification(testEmail), oldNotification)
	n, s := Strategy(nil, a)

	if n == nil {
		t.Fatal("expected a notification, got nil")
	}
	if n.Email != testEmail {
		t.Fatalf("incorrect aggregate email: got %s, wanted %s", n.Email, testEmail)
	}
	if l := len(n.Notifications); l != 2 {
		t.Fatalf("expected 3 messages in the aggregation, got %d", l)
	}

	if s != nil {
		t.Fatalf("got aggregation state: %#v", s)
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
