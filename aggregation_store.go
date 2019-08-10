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

func (a *AggregateStore) Get(id string) ([]*SecurityNotification, error) {
	var sns []*SecurityNotification
	err := a.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &sns)
		})
	})

	if err == badger.ErrKeyNotFound {
		return nil, nil
	}

	return sns, nil
}

func (a *AggregateStore) Save(id string, state []*SecurityNotification) error {
	b, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return a.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(id), b)
	})
}
