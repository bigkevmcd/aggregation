package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dgraph-io/badger"
)

func main() {
	db, err := badger.Open(badger.DefaultOptions("./tmp"))
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	store := NewStore(db)

	http.HandleFunc("/notifications", makeHandler(store))

	fmt.Printf("receiving on http://localhost:8080/notifications\n")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func makeHandler(store *AggregateStore) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var sn SecurityNotification
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&sn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		correlationID := sn.Email
		existingState, err := store.Get(correlationID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		n, s := Strategy(&sn, existingState)
		err = store.Save(correlationID, s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if n != nil {
			log.Printf("new event emitted: %#v\n", n)
		}

	}
}