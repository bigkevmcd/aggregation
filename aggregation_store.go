package aggregator

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger"
)

var defaultPrefix = "aggregator"

type AggregationProcessor func(state Aggregation) (*AggregateNotification, Aggregation)
type StrategyFunc func(*SecurityNotification, Aggregation) (*AggregateNotification, Aggregation)

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
		item, err := txn.Get(keyForId(defaultPrefix, id))
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

func (a *AggregateStore) Save(id string, state Aggregation) error {
	b, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return a.db.Update(func(txn *badger.Txn) error {
		return txn.Set(keyForId(defaultPrefix, id), b)
	})
}

func (a *AggregateStore) ProcessAggregates(f AggregationProcessor) error {
	err := a.db.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(defaultPrefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			itemKey := item.Key()
			err := item.Value(func(v []byte) error {
				var sns Aggregation
				if err := json.Unmarshal(v, &sns); err != nil {
					return err
				}
				// TODO: send a notification
				_, newState := f(sns)

				if newState == nil {
					txn.Delete(itemKey)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (a *AggregateStore) ExecuteAggregation(n *SecurityNotification, f StrategyFunc, p Publisher) error {
	return a.db.Update(func(txn *badger.Txn) error {
		id := n.Email
		previous, err := getOrEmpty(txn, id)
		if err != nil {
			return err
		}
		n, newState := f(n, previous)
		err = p.Publish(n)
		if err != nil {
			return err
		}
		b, err := json.Marshal(newState)
		if err != nil {
			return err
		}
		return txn.Set(keyForId(defaultPrefix, id), b)
	})
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
		return json.Unmarshal(val, &sns)
	})
	return sns, err
}
