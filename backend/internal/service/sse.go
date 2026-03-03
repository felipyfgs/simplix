package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/simplix/api/internal/domain"
)

// SSEBroker manages Server-Sent Events connections.
type SSEBroker struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]map[string]chan domain.SSEEvent // userID -> clientID -> chan
}

func NewSSEBroker() *SSEBroker {
	return &SSEBroker{clients: make(map[uuid.UUID]map[string]chan domain.SSEEvent)}
}

// Subscribe registers a channel for the given user and returns a unique client ID and channel.
func (b *SSEBroker) Subscribe(userID uuid.UUID) (string, <-chan domain.SSEEvent) {
	clientID := uuid.New().String()
	ch := make(chan domain.SSEEvent, 32)
	b.mu.Lock()
	if b.clients[userID] == nil {
		b.clients[userID] = make(map[string]chan domain.SSEEvent)
	}
	b.clients[userID][clientID] = ch
	b.mu.Unlock()
	return clientID, ch
}

// Unsubscribe removes the client channel.
func (b *SSEBroker) Unsubscribe(userID uuid.UUID, clientID string) {
	b.mu.Lock()
	if m, ok := b.clients[userID]; ok {
		if ch, ok := m[clientID]; ok {
			close(ch)
			delete(m, clientID)
		}
		if len(m) == 0 {
			delete(b.clients, userID)
		}
	}
	b.mu.Unlock()
}

// Broadcast sends an event to all connected clients of the given user.
// If userID is the zero UUID, broadcast to ALL connected users.
func (b *SSEBroker) Broadcast(userID uuid.UUID, evt domain.SSEEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	zero := uuid.UUID{}
	for uid, clients := range b.clients {
		if userID != zero && uid != userID {
			continue
		}
		for _, ch := range clients {
			select {
			case ch <- evt:
			default: // drop if buffer full
			}
		}
	}
}

// BroadcastAll sends to every connected client regardless of user.
func (b *SSEBroker) BroadcastAll(evt domain.SSEEvent) {
	b.Broadcast(uuid.UUID{}, evt)
}

// ServeHTTP is an http.HandlerFunc that streams SSE to the connected client.
func (b *SSEBroker) ServeSSE(userID uuid.UUID, w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	clientID, ch := b.Subscribe(userID)
	defer b.Unsubscribe(userID, clientID)

	// Send initial ping
	fmt.Fprintf(w, "event: ping\ndata: {}\n\n")
	flusher.Flush()

	for {
		select {
		case evt, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(evt.Payload)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
