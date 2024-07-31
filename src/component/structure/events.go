package structure

import (
	"slices"
	"sync"
)

type EventType byte

const NewDirectedRelationEventType byte = 0

type Event interface {
	GetType() EventType
	GetGraphID() GraphID
	GetData() any
}

type NewDirectedRelationEventData struct {
	DirectedRelation DirectedRelation
	GraphRelevance   float32
}

type NewDirectedRelationEvent struct {
	NewDirectedRelationEventData NewDirectedRelationEventData
	GraphID                      GraphID
}

func (e NewDirectedRelationEvent) GetType() EventType {
	return EventType(NewDirectedRelationEventType)
}

func (e NewDirectedRelationEvent) GetGraphID() GraphID {
	return e.GraphID
}

func (e NewDirectedRelationEvent) GetData() any {
	return e.NewDirectedRelationEventData
}

type EventManager struct {
	subscribers      map[GraphID][]*func(Event)
	events           chan Event
	subscribersMutex *sync.RWMutex
}

func NewEventManager() *EventManager {
	eventManager := &EventManager{
		subscribersMutex: &sync.RWMutex{},
		subscribers:      make(map[GraphID][]*func(Event)),
		events:           make(chan Event),
	}

	go func() {
		for {
			event := <-eventManager.events
			eventManager.subscribersMutex.RLock()
			subscribers := eventManager.subscribers[event.GetGraphID()]
			eventManager.subscribersMutex.RUnlock()
			for _, subscriber := range subscribers {
				(*subscriber)(event)
			}
		}
	}()

	return eventManager
}

func (e *EventManager) Subscribe(graphId GraphID, callback *func(Event)) {
	e.subscribersMutex.Lock()
	defer e.subscribersMutex.Unlock()
	e.subscribers[graphId] = append(e.subscribers[graphId], callback)
}

func (e *EventManager) Unsubscribe(graphId GraphID, callback *func(Event)) {
	e.subscribersMutex.Lock()
	defer e.subscribersMutex.Unlock()
	index := -1
	for i, existingCallback := range e.subscribers[graphId] {
		if existingCallback == callback {
			index = i
			break
		}
	}

	if index >= 0 {
		e.subscribers[graphId] = slices.Delete(e.subscribers[graphId], index, index+1)
	}

}

func (e *EventManager) Publish(graphID GraphID, event Event) {
	e.events <- event
}
