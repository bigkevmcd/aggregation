package aggregator

type Publisher interface {
	Publish(*AggregateNotification) error
}

// Executed within the context of a transaction, receives the incoming event and
// the current aggregate state for that event, returns the new state and an
// error if any.
type Processor interface {
	Process(*SecurityNotification, Aggregation) (Aggregation, error)
}
