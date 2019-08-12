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

func TestProcessNotificationWithNoPreviousState(t *testing.T) {
	store, cleanup := createBadgerStore(t)
	defer cleanup()

	processor := &mockProcessor{}
	notification := makeNotification(testEmail)

	err := store.ProcessNotification(notification, processor)
	fatalIfError(t, err)

	if processor.processedNotification == nil {
		t.Fatalf("processor did not receive notification")
	}
}

func TestProcessNotificationUpdatesExistingState(t *testing.T) {
	store, cleanup := createBadgerStore(t)
	defer cleanup()

	notifications := Aggregation{makeNotification(testEmail)}
	err := store.Save(testEmail, notifications)
	fatalIfError(t, err)
	processor := &mockProcessor{}
	notification := makeNotification(testEmail)
	processor.returnAggregation = Aggregation{makeNotification(testEmail), makeNotification(testEmail)}

	err = store.ProcessNotification(notification, processor)

	fatalIfError(t, err)
	if processor.processedNotification == nil {
		t.Fatalf("processor did not receive notification")
	}

	loaded, err := store.Get(testEmail)
	fatalIfError(t, err)
	if !reflect.DeepEqual(processor.returnAggregation, loaded) {
		t.Fatalf("got %#v, wanted %#v", loaded, processor.returnAggregation)
	}
}

func TestExecuteAggregationErrorPublishing(t *testing.T) {
	store, cleanup := createBadgerStore(t)
	defer cleanup()
	testError := errors.New("this is a test")

	processor := &mockProcessor{}
	processor.err = testError
	notification := makeNotification(testEmail)

	err := store.ProcessNotification(notification, processor)
	if err != testError {
		t.Fatalf("got error %s, wanted %s", err, testError)
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

type mockProcessor struct {
	processedNotification *SecurityNotification
	processedAggregation  Aggregation
	returnAggregation     Aggregation
	err                   error
}

func (p *mockProcessor) Process(n *SecurityNotification, a Aggregation) (Aggregation, error) {
	p.processedNotification = n
	p.processedAggregation = a
	return p.returnAggregation, p.err
}
