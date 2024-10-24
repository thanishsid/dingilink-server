package messaging

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/nats-io/nats.go"
)

// Client struct represents a WebSocket connection
type Client[V any] struct {
	id      string
	sub     *nats.Subscription
	channel chan V
}

// ChannelManager holds the map of user ID to channels for subscriptions
type ChannelManager[V any] struct {
	natsConn *nats.Conn
	clients  map[string]*Client[V]
	mu       sync.RWMutex
}

func NewChannelManager[V any](natsURL string) (*ChannelManager[V], error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}

	return &ChannelManager[V]{
		natsConn: nc,
		clients:  make(map[string]*Client[V]),
	}, nil
}

// Subscribe adds a value to the channel map
func (m *ChannelManager[V]) Subscribe(id string) (chan V, error) {
	ch := make(chan V)

	// Subscribe to NATS for this user's subject
	sub, err := m.natsConn.Subscribe(id, func(msg *nats.Msg) {
		var payload V

		if err := json.Unmarshal(msg.Data, &payload); err != nil {
			log.Println("Failed to deserialize message:", err)
			return
		}

		ch <- payload
	})
	if err != nil {
		return nil, err
	}

	client := Client[V]{
		id:      id,
		sub:     sub,
		channel: ch,
	}

	m.mu.Lock()
	m.clients[id] = &client
	m.mu.Unlock()

	return ch, nil
}

// Unsubscribe a client
func (m *ChannelManager[V]) Unsubscribe(id string) error {
	c, exists := m.clients[id]
	if !exists {
		return nil
	}

	if err := c.sub.Unsubscribe(); err != nil {
		return err
	}

	m.mu.Lock()
	delete(m.clients, id)
	m.mu.Unlock()

	return nil
}

// Send an payload to the id and subject.
func (m *ChannelManager[V]) SendPayload(id string, payload V) error {
	m.mu.RLock() // Lock before accessing the channels map
	defer m.mu.RUnlock()

	client, exists := m.clients[id]
	if exists {
		client.channel <- payload
	} else {
		payloadJson, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		if err := m.natsConn.Publish(id, payloadJson); err != nil {
			return err
		}
	}

	return nil
}
