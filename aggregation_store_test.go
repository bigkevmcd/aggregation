package main

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/dgraph-io/badger"
)

const testEmail = "a@example.com"

func TestGetWithUnknownID(t *testing.T) {
	store, cleanup := createBadger(t)
	defer cleanup()

	a, err := store.Get(testEmail)
	fatalIfError(t, err)
	if a != nil {
		t.Fatalf("unknown ID wanted %#v, got %#v", nil, a)
	}
}

func TestSave(t *testing.T) {
	store, cleanup := createBadger(t)
	defer cleanup()

	notifications := []*SecurityNotification{makeNotification(testEmail)}
	an := &AggregateNotification{
		Notifications: notifications,
	}

	err := store.Save(testEmail, an)
	fatalIfError(t, err)

	loaded, err := store.Get(testEmail)
	fatalIfError(t, err)
	if !reflect.DeepEqual(an, loaded) {
		t.Fatalf("save failed to save: wanted %#v, got %#v", an, loaded)
	}
}

func createBadger(t *testing.T) (*AggregateStore, func()) {
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
