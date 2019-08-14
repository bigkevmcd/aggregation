package aggregator

import (
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

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

	agg := makeAggregation(makeNotification(testEmail))
	err := store.Save(testEmail, agg)
	fatalIfError(t, err)

	loaded, err := store.Get(testEmail)
	fatalIfError(t, err)
	if !reflect.DeepEqual(agg, loaded) {
		t.Fatalf("save failed to save: wanted %#v, got %#v", agg, loaded)
	}
}

func TestSaveUpdatesLastSaved(t *testing.T) {
	store, cleanup := createBadgerStore(t)
	defer cleanup()
	oldClock := clock
	defer func() {
		clock = oldClock
	}()
	lastUpdated := time.Now().UTC()
	clock = func() time.Time {
		return lastUpdated
	}

	agg := makeAggregation(makeNotification(testEmail))

	err := store.Save(testEmail, agg)
	fatalIfError(t, err)

	loaded, err := store.Get(testEmail)
	fatalIfError(t, err)
	if !loaded.LastSaved.Equal(lastUpdated) {
		t.Fatalf("last updated field not changed: got %v, wanted %v", loaded.LastSaved, lastUpdated)
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

	agg := makeAggregation(makeNotification(testEmail))
	err := store.Save(testEmail, agg)
	fatalIfError(t, err)
	processor := &mockProcessor{}
	notification := makeNotification(testEmail)
	processor.returnAggregation = makeAggregation(makeNotification(testEmail), makeNotification(testEmail))

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
		t.Fatalf("saved aggregation despite error: %#v", loaded)
	}
}

func TestProcessAggregationsWithNoAggregations(t *testing.T) {
	store, cleanup := createBadgerStore(t)
	defer cleanup()
	processor := &mockAggregationProcessor{}

	err := store.ProcessAggregations(processor)
	fatalIfError(t, err)

	if processor.count != 0 {
		t.Fatalf("processed aggregations: got %d, wanted 0", processor.count)
	}
}

func TestProcessAggregationsWithAnAggregate(t *testing.T) {
	store, cleanup := createBadgerStore(t)
	defer cleanup()
	agg := makeAggregation(makeNotification(testEmail))
	err := store.Save(testEmail, agg)
	fatalIfError(t, err)
	processor := &mockAggregationProcessor{}

	err = store.ProcessAggregations(processor)
	fatalIfError(t, err)

	if processor.count != 1 {
		t.Fatalf("processed aggregations: got %d, wanted 1", processor.count)
	}
}

func TestProcessAggregationsCleansUpAggregations(t *testing.T) {
	store, cleanup := createBadgerStore(t)
	defer cleanup()
	agg := makeAggregation(makeNotification(testEmail))
	err := store.Save(testEmail, agg)
	fatalIfError(t, err)
	processor1 := &mockAggregationProcessor{}
	err = store.ProcessAggregations(processor1)
	fatalIfError(t, err)

	processor2 := &mockAggregationProcessor{}
	err = store.ProcessAggregations(processor2)
	fatalIfError(t, err)

	if processor2.count != 0 {
		t.Fatalf("reprocessed aggregations: got %d, wanted 0", processor2.count)
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

func makeAggregation(n ...*SecurityNotification) *Aggregation {
	return &Aggregation{
		Email:         testEmail,
		LastSaved:     time.Now().UTC(),
		Notifications: n,
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
	processedAggregation  *Aggregation
	returnAggregation     *Aggregation
	err                   error
}

func (p *mockProcessor) Process(n *SecurityNotification, a *Aggregation) (*Aggregation, error) {
	p.processedNotification = n
	p.processedAggregation = a
	return p.returnAggregation, p.err
}

type mockAggregationProcessor struct {
	count int32
	err   error
}

func (p *mockAggregationProcessor) Process(a *Aggregation) (*Aggregation, error) {
	p.count++
	return nil, p.err
}
