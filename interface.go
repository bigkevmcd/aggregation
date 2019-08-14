package aggregator

// Publisher is responsible for taking an AggregateNotification and sending it
// to a final destination, this final destination should guarantee at least once
// delivery.
type Publisher interface {
	Publish(*AggregateNotification) error
}

// Processor handles incoming notifications and previous state.
// Executed within the context of a transaction, receives the incoming event and
// the current aggregate state for that event, returns the new state and an
// error if any.
type Processor interface {
	Process(*SecurityNotification, *Aggregation) (*Aggregation, error)
}

// AggregationProcessor handles bulk aggregation processing.
//
// There is no new event in this case, if it returns nil, then the existing
// state should be removed from the aggregation store.
type AggregationProcessor interface {
	ProcessWithoutEvent(*Aggregation) (*Aggregation, error)
}
