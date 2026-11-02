package structure

import "sync"

var subscriberHandleCounter int = 0

type GraphAddEdgeEvent[T Stage] struct {
	GraphID  GraphID
	Relation *DirectedRelation[T]
}

type EventManager[T Stage] struct {
	subscribers      map[GraphID]map[int]chan *GraphAddEdgeEvent[T]
	subscribersMutex *sync.RWMutex
}

type SubscriberHandle struct {
	ID      int
	GraphID GraphID
}

func NewEventManager[T Stage]() *EventManager[T] {
	eventManager := &EventManager[T]{
		subscribers:      map[GraphID]map[int]chan *GraphAddEdgeEvent[T]{},
		subscribersMutex: &sync.RWMutex{},
	}

	return eventManager
}

func (eventManager *EventManager[T]) Subscribe(graphId GraphID, subscriber chan *GraphAddEdgeEvent[T]) SubscriberHandle {
	eventManager.subscribersMutex.Lock()
	_, ok := eventManager.subscribers[graphId]
	if !ok {
		eventManager.subscribers[graphId] = map[int]chan *GraphAddEdgeEvent[T]{}
	}

	id := subscriberHandleCounter
	subscriberHandleCounter += 1
	eventManager.subscribers[graphId][id] = subscriber

	eventManager.subscribersMutex.Unlock()

	return SubscriberHandle{
		ID:      id,
		GraphID: graphId,
	}
}

func (eventManager *EventManager[T]) Unsubscribe(handle SubscriberHandle) {
	eventManager.subscribersMutex.Lock()
	defer eventManager.subscribersMutex.Unlock()
	_, ok := eventManager.subscribers[handle.GraphID]
	if !ok {
		return
	}

	delete(eventManager.subscribers[handle.GraphID], handle.ID)
	if len(eventManager.subscribers[handle.GraphID]) == 0 {
		delete(eventManager.subscribers, handle.GraphID)
	}
}

func (eventManager *EventManager[T]) Publish(graphId GraphID, relation *DirectedRelation[T]) {
	event := GraphAddEdgeEvent[T]{
		GraphID:  graphId,
		Relation: relation,
	}
	eventManager.subscribersMutex.RLock()
	for _, subscriber := range eventManager.subscribers[event.GraphID] {
		select {
		case subscriber <- &event:
		default: // Drop event
		}
	}
	eventManager.subscribersMutex.RUnlock()
}
