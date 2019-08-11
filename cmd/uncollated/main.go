package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bigkevmcd/aggregator"
)

var existingState aggregator.Aggregation

func main() {
	existingState = make(aggregator.Aggregation, 0)
	http.HandleFunc("/notifications", aggregatorHandler)

	fmt.Printf("receiving on http://localhost:8080/notifications\n")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func aggregatorHandler(w http.ResponseWriter, r *http.Request) {
	var sn aggregator.SecurityNotification
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&sn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var n *aggregator.AggregateNotification
	n, existingState = aggregator.Strategy(&sn, existingState)
	if n != nil {
		log.Printf("new event emitted for user %s\n", n.Email)
		return
	}
	log.Println("event processed - no event emitted")
}
