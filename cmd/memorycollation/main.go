package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bigkevmcd/aggregator"
)

var aggregationStore map[string]aggregator.Aggregation

func main() {
	aggregationStore = make(map[string]aggregator.Aggregation)
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

	correlationID := sn.Email
	log.Printf("processing %s priority event for %s\n", aggregator.TranslatePriority(sn.Priority), correlationID)

	existingState, ok := aggregationStore[correlationID]
	if !ok {
		existingState = make(aggregator.Aggregation, 0)
	}

	n, newState := aggregator.Strategy(&sn, existingState)
	aggregationStore[correlationID] = newState

	if n != nil {
		log.Printf("new event emitted for user %s\n", n.Email)
		return
	}
	log.Println("event processed - no event emitted")
}
