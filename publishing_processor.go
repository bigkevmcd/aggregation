package aggregator

import "log"

type PublishingProcessor struct {
}

func (p *PublishingProcessor) Process(evt *SecurityNotification, existingState *Aggregation) (*Aggregation, error) {
	notification, newState := Strategy(evt, existingState)
	if notification == nil {
		return newState, nil
	}
	log.Printf("publishing %#v\n", notification)
	return newState, nil

}

func (p *PublishingProcessor) ProcessWithoutEvent(existingState *Aggregation) (*Aggregation, error) {
	notification, newState := StrategyWithoutEvent(existingState)
	if notification == nil {
		return newState, nil
	}
	log.Printf("publishing %#v\n", notification)
	return newState, nil
}
