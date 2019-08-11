package aggregator

type Publisher interface {
	Publish(*AggregateNotification) error
}
