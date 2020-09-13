package events

import (
	"encoding/json"
	"fmt"
	"github.com/gammazero/workerpool"
	"github.com/pkg/errors"
	"strings"
	"sync"
)

type Event struct {
	Data  string
	Topic string
}

type EventBus struct {
	mu    sync.RWMutex
	pools map[string]*CallbackPool
}

func New() *EventBus {
	return &EventBus{
		pools: make(map[string]*CallbackPool),
	}
}

// Publish data to a given topic.
func (e *EventBus) Publish(topic string, data string) {
	t := topic
	// Some of our topics for the socket support passing a more specific namespace,
	// such as "backup completed:1234" to indicate which specific backup was completed.
	//
	// In these cases, we still need to the send the event using the standard listener
	// name of "backup completed".
	if strings.Contains(topic, ":") {
		parts := strings.SplitN(topic, ":", 2)

		if len(parts) == 2 {
			t = parts[0]
		}
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	// Acquire a read lock and loop over all of the channels registered for the topic. This
	// avoids a panic crash if the process tries to unregister the channel while this routine
	// is running.
	if cp, ok := e.pools[t]; ok {
		evt := Event{Data: data, Topic: topic}

		for _, callback := range cp.callbacks {
			c := *callback
			evt := evt
			// Using the workerpool with one worker allows us to execute events in a FIFO manner. Running
			// this using goroutines would cause things such as console output to just output in random order
			// if more than one event is fired at the same time.
			//
			// However, the pool submission does not block the execution of this function itself, allowing
			// us to call publish without blocking any of the other pathways.
			//
			// @see https://github.com/pterodactyl/panel/issues/2303
			fmt.Println("pre-submit for topic:", evt.Topic)
			cp.pool.Submit(func() {
				fmt.Println("executing callback for event:", evt.Topic)
				c(evt)
			})
		}
	}
}

// Publishes a JSON message to a given topic.
func (e *EventBus) PublishJson(topic string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return errors.WithStack(err)
	}

	e.Publish(topic, string(b))

	return nil
}

// Register a callback function that will be executed each time one of the events using the topic
// name is called.
func (e *EventBus) On(topic string, callback *func(Event)) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if this topic has been registered at least once for the event listener, and if
	// not create an empty struct for the topic.
	if _, exists := e.pools[topic]; !exists {
		fmt.Println("no pool for topic, creating:", topic)
		e.pools[topic] = &CallbackPool{
			callbacks: make([]*func(Event), 0),
			pool:      workerpool.New(1),
		}
	}

	// If this callback is not already registered as an event listener, go ahead and append
	// it to the array of callbacks for this topic.
	e.pools[topic].Add(callback)
}

// Removes an event listener from the bus.
func (e *EventBus) Off(topic string, callback *func(Event)) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if cp, ok := e.pools[topic]; ok {
		cp.Remove(callback)
	}
}

// Removes all of the event listeners that have been registered for any topic. Also stops the worker
// pool to close that routine.
func (e *EventBus) Destroy() {
	e.mu.Lock()
	defer e.mu.Unlock()

	fmt.Println("destroying pool for event bus")
	// Stop every pool that exists for a given callback topic.
	for _, cp := range e.pools {
		cp.pool.Stop()
	}

	e.pools = make(map[string]*CallbackPool)
}
