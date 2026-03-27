package chat

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewHub(t *testing.T) {
	h := NewHub()
	if h.clients == nil {
		t.Error("Hub clients map is nil")
	}
	if h.rooms == nil {
		t.Error("Hub rooms map is nil")
	}
}

func TestJoinRoom(t *testing.T) {
	h := NewHub()
	client := &Client{
		Hub: h,
		ID:  uuid.New().String(),
		// Connection intentionally left nil for unit tests not involving I/O
	}
	roomID := "test-room"

	err := h.joinRoom(client, roomID)
	if err != nil {
		t.Errorf("joinRoom failed: %v", err)
	}

	h.mu.RLock()
	room, ok := h.rooms[roomID]
	h.mu.RUnlock()

	if !ok {
		t.Errorf("Room %s was not created", roomID)
	}

	room.mu.RLock()
	if !room.Clients[client] {
		t.Errorf("Client was not added to room %s", roomID)
	}
	room.mu.RUnlock()
}

func TestLeaveRoom(t *testing.T) {
	h := NewHub()
	client := &Client{
		Hub: h,
		ID:  uuid.New().String(),
	}
	roomID := "test-room"

	h.joinRoom(client, roomID)
	h.leaveRoom(client)

	h.mu.RLock()
	_, ok := h.rooms[roomID]
	h.mu.RUnlock()

	if ok {
		t.Errorf("Room %s should have been deleted when empty", roomID)
	}
}

func TestBroadcast(t *testing.T) {
	h := NewHub()
	roomID := "test-room"

	// Mock client with channel
	client := &Client{
		Hub:  h,
		ID:   "tester",
		send: make(chan Message, 1),
	}

	h.joinRoom(client, roomID)

	msg := Message{
		Sender:  1,
		Content: "test message",
	}

	h.Broadcast(roomID, msg)

	select {
	case receivedMsg := <-client.send:
		if receivedMsg.Content != msg.Content {
			t.Errorf("Expected message content %s, got %s", msg.Content, receivedMsg.Content)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Broadcast message was not received")
	}
}

func TestBroadcastJSON(t *testing.T) {
	h := NewHub()
	roomID := "test-room"

	client := &Client{
		Hub:  h,
		ID:   "tester",
		send: make(chan Message, 1),
	}

	h.joinRoom(client, roomID)

	jsonContent := map[string]interface{}{
		"key":   "value",
		"count": 42,
	}

	msg := Message{
		Sender:  1,
		Content: jsonContent,
	}

	h.Broadcast(roomID, msg)

	select {
	case receivedMsg := <-client.send:
		receivedMap, ok := receivedMsg.Content.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected content to be a map, got %T", receivedMsg.Content)
		}
		if receivedMap["key"] != "value" || receivedMap["count"] != 42 {
			t.Errorf("Expected content %v, got %v", jsonContent, receivedMsg.Content)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Broadcast message was not received")
	}
}
