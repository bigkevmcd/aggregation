package aggregator

import (
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/dgraph-io/badger"
)

func TestGetWithUnknownID(t *testing.T) {
	store, cleanup := createBadgerStore(t)
	defer cleanup()

	a, err := store.Get(testEmail)
	fatalIfError(t, err)
	if a != nil {
		t.Fatalf("unknown ID wanted %#v, got %#v", nil, a)
	}
}

func TestSave(t *testing.T) {
	store, cleanup := createBadgerStore(t)
	defer cleanup()

	notifications := Aggregation{makeNotification(testEmail)}
	err := store.Save(testEmail, notifications)
	fatalIfError(t, err)

	loaded, err := store.Get(testEmail)
	fatalIfError(t, err)
	if !reflect.DeepEqual(notifications, loaded) {
		t.Fatalf("save failed to save: wanted %#v, got %#v", notifications, loaded)
	}
}

func TestProcessAggregatesWithNoAggregates(t *testing.T) {
	store, cleanup := createBadgerStore(t)
	defer cleanup()

	count := 0
	err := store.ProcessAggregates(func(state Aggregation) (*AggregateNotification, Aggregation) {
		count++
		return nil, nil
	})

	fatalIfError(t, err)
	if count != 0 {
		t.Fatalf("processed aggregates: got %d, wanted 0", count)
	}
}

func TestProcessAggregatesWithAnAggregate(t *testing.T) {
	store, cleanup := createBadgerStore(t)
	defer cleanup()
	notifications := Aggregation{makeNotification(testEmail)}
	err := store.Save(testEmail, notifications)
	fatalIfError(t, err)

	count := 0
	err = store.ProcessAggregates(func(state Aggregation) (*AggregateNotification, Aggregation) {
		count++
		return nil, nil
	})
	fatalIfError(t, err)

	if count != 1 {
		t.Fatalf("processed aggregates: got %d, wanted 1", count)
	}
}

func TestProcessAggregatesSavesNewAggregates(t *testing.T) {
	t.Skip()
}

func TestProcessAggregatesCleansUpAggregates(t *testing.T) {
	store, cleanup := createBadgerStore(t)
	defer cleanup()
	notifications := Aggregation{makeNotification(testEmail)}
	err := store.Save(testEmail, notifications)
	fatalIfError(t, err)

	count := 0
	err = store.ProcessAggregates(func(state Aggregation) (*AggregateNotification, Aggregation) {
		count++
		return nil, nil
	})
	fatalIfError(t, err)

	newCount := 0
	err = store.ProcessAggregates(func(state Aggregation) (*AggregateNotification, Aggregation) {
		newCount++
		return nil, nil
	})
	fatalIfError(t, err)

	if newCount != 0 {
		t.Fatalf("reprocessed aggregates: got %d, wanted 0", newCount)
	}
}

func TestExecuteAggregationWithUnknownCorrelationID(t *testing.T) {
	store, cleanup := createBadgerStore(t)
	defer cleanup()

	publisher := &mockPublisher{}
	notification := makeNotification(testEmail)

	err := store.ExecuteAggregation(notification, Strategy, publisher)
	fatalIfError(t, err)
	if l := len(publisher.published); l != 1 {
		t.Fatalf("publishing notifications: got %d, wanted 1", l)
	}
}

func TestExecuteAggregationErrorPublishing(t *testing.T) {
	store, cleanup := createBadgerStore(t)
	defer cleanup()
	testError := errors.New("this is a test")

	publisher := &mockPublisher{}
	publisher.err = testError
	notification := makeNotification(testEmail)

	err := store.ExecuteAggregation(notification, Strategy, publisher)
	if err != testError {
		t.Fatalf("got error %s, wanted %s", err, testError)
	}
	if l := len(publisher.published); l != 1 {
		t.Fatalf("publishing notifications: got %d, wanted 1", l)
	}
	loaded, err := store.Get(testEmail)
	fatalIfError(t, err)
	if loaded != nil {
		t.Fatalf("saved aggregate despite error: %#v", loaded)
	}
}

func createBadgerStore(t *testing.T) (*AggregateStore, func()) {
	dir, err := ioutil.TempDir(os.TempDir(), "badger")
	if err != nil {
		t.Fatal(err)
	}
	db, err := badger.Open(badger.DefaultOptions(dir))
	if err != nil {
		t.Fatal(err)
	}

	return NewStore(db), func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

func fatalIfError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

type mockPublisher struct {
	published []*AggregateNotification
	err       error
}

func (m *mockPublisher) Publish(n *AggregateNotification) error {
	m.published = append(m.published, n)
	return m.err
}
