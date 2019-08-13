package aggregator

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger"
)

var defaultPrefix = "aggregator"

type AggregateStore struct {
	db *badger.DB
}

func NewStore(db *badger.DB) *AggregateStore {
	return &AggregateStore{
		db: db,
	}
}

func (a *AggregateStore) Get(id string) (Aggregation, error) {
	var sns Aggregation
	err := a.db.View(func(txn *badger.Txn) error {
		var err error
		sns, err = getOrEmpty(txn, id)
		return err
	})
	return sns, err
}

func (a *AggregateStore) Save(id string, state Aggregation) error {
	b, err := marshal(state)
	if err != nil {
		return err
	}
	return a.db.Update(func(txn *badger.Txn) error {
		return txn.Set(keyForId(defaultPrefix, id), b)
	})
}

func (a *AggregateStore) ProcessNotification(n *SecurityNotification, p Processor) error {
	return a.db.Update(func(txn *badger.Txn) error {
		id := n.Email
		previous, err := getOrEmpty(txn, id)
		if err != nil {
			return err
		}
		newState, err := p.Process(n, previous)
		if err != nil {
			return err
		}
		b, err := marshal(newState)
		if err != nil {
			return err
		}
		return txn.Set(keyForId(defaultPrefix, id), b)
	})
}

func (a *AggregateStore) ProcessAggregations(p AggregationProcessor) error {
	err := a.db.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(defaultPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := processItem(txn, item, item.Key(), p)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func keyForId(prefix, id string) []byte {
	return []byte(fmt.Sprintf("%s:%s", prefix, id))
}

func getOrEmpty(txn *badger.Txn, id string) (Aggregation, error) {
	item, err := txn.Get(keyForId(defaultPrefix, id))
	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var sns Aggregation
	err = item.Value(func(val []byte) error {
		sns, err = unmarshal(val)
		return err
	})
	return sns, err
}

func marshal(a Aggregation) ([]byte, error) {
	return json.Marshal(a)
}

func unmarshal(b []byte) (Aggregation, error) {
	var sns Aggregation
	if err := json.Unmarshal(b, &sns); err != nil {
		return nil, err
	}
	return sns, nil
}

func processItem(txn *badger.Txn, item *badger.Item, key []byte, p AggregationProcessor) error {
	err := item.Value(func(v []byte) error {
		state, err := unmarshal(v)
		if err != nil {
			return err
		}
		newState, err := p.Process(state)
		if err != nil {
			return err
		}
		if newState == nil {
			txn.Delete(key)
		}
		return nil
	})
	return err
}
