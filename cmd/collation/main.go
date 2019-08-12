package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dgraph-io/badger"

	"github.com/bigkevmcd/aggregator"
)

func main() {
	db, err := badger.Open(badger.DefaultOptions("./tmp"))
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	store := aggregator.NewStore(db)
	publisher := &aggregator.PublishingProcessor{}

	http.HandleFunc("/notifications", makeHandler(store, publisher))

	fmt.Printf("receiving on http://localhost:8080/notifications\n")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func makeHandler(store *aggregator.AggregateStore, processor aggregator.Processor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var sn aggregator.SecurityNotification
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&sn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = store.ProcessNotification(&sn, processor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
