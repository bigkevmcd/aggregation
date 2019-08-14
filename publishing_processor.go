package aggregator

import "log"

type PublishingProcessor struct {
}

func (p *PublishingProcessor) Process(evt *SecurityNotification, oldState *Aggregation) (*Aggregation, error) {
	notification, newState := Strategy(evt, oldState)
	if notification == nil {
		return newState, nil
	}
	log.Printf("publishing %#v\n", notification)
	return newState, nil

}
