package main

import (
	"encoding/json"

	"github.com/dgraph-io/badger"
)

type AggregateStore struct {
	db *badger.DB
}

func NewStore(db *badger.DB) *AggregateStore {
	return &AggregateStore{
		db: db,
	}
}

func (a *AggregateStore) Get(id string) (*AggregateNotification, error) {
	var an AggregateNotification
	err := a.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &an)
		})
	})

	if err == badger.ErrKeyNotFound {
		return nil, nil
	}

	return &an, nil
}

func (a *AggregateStore) Save(id string, an *AggregateNotification) error {
	b, err := json.Marshal(an)
	if err != nil {
		return err
	}
	return a.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(id), b)
	})
}
