package websocket

import (
	"encoding/json"
	"sync"
)

// Hub owns the in-memory index of online students. Its mutex also protects
// Client.send from being closed while a sender is queuing an outbound event.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]*Client)}
}

// Register installs one active socket per student; duplicate connections are
// rejected so a newer socket cannot silently steal the student's chat stream.
func (hub *Hub) Register(client *Client) bool {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	key := client.StudentID.Hex()
	if _, exists := hub.clients[key]; exists {
		return false
	}
	hub.clients[key] = client
	return true
}

func (hub *Hub) Unregister(client *Client) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	key := client.StudentID.Hex()
	if hub.clients[key] != client {
		return
	}
	delete(hub.clients, key)
	close(client.send)
}

// RouteMessage queues data for one online student without blocking a read pump.
func (hub *Hub) RouteMessage(studentID string, payload []byte) bool {
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	client, exists := hub.clients[studentID]
	if !exists {
		return false
	}
	select {
	case client.send <- payload:
		return true
	default:
		return false
	}
}

// NotifyMatchFound pushes a MATCH_FOUND event to each connected participant.
// The return value is the number of participants that received a queued event.
func (hub *Hub) NotifyMatchFound(student1ID, student2ID string, matchData interface{}) int {
	payload, err := json.Marshal(struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{Type: "MATCH_FOUND", Data: matchData})
	if err != nil {
		return 0
	}

	delivered := 0
	if hub.RouteMessage(student1ID, payload) {
		delivered++
	}
	if hub.RouteMessage(student2ID, payload) {
		delivered++
	}
	return delivered
}

func (hub *Hub) Online(studentID string) bool {
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	_, exists := hub.clients[studentID]
	return exists
}
