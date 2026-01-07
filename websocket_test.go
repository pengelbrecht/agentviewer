package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

// TestNewHub tests the creation of a new Hub.
func TestNewHub(t *testing.T) {
	hub := NewHub()

	if hub == nil {
		t.Fatal("NewHub returned nil")
	}
	if hub.clients == nil {
		t.Error("clients map is nil")
	}
	if hub.register == nil {
		t.Error("register channel is nil")
	}
	if hub.unregister == nil {
		t.Error("unregister channel is nil")
	}
	if hub.broadcast == nil {
		t.Error("broadcast channel is nil")
	}
	if hub.done == nil {
		t.Error("done channel is nil")
	}
	if hub.ClientCount() != 0 {
		t.Errorf("expected 0 clients, got %d", hub.ClientCount())
	}
}

// TestHubClientCount tests the ClientCount method.
func TestHubClientCount(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	if hub.ClientCount() != 0 {
		t.Errorf("expected 0 clients initially, got %d", hub.ClientCount())
	}
}

// TestHubRegisterUnregister tests client registration and unregistration.
func TestHubRegisterUnregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	// Create a mock client (no real WebSocket connection needed for this test)
	client := &Client{
		hub:  hub,
		conn: nil, // We won't use the connection in this test
		send: make(chan []byte, 256),
	}

	// Register the client
	hub.register <- client
	time.Sleep(10 * time.Millisecond) // Give goroutine time to process

	if hub.ClientCount() != 1 {
		t.Errorf("expected 1 client after register, got %d", hub.ClientCount())
	}

	// Unregister the client
	hub.unregister <- client
	time.Sleep(10 * time.Millisecond) // Give goroutine time to process

	if hub.ClientCount() != 0 {
		t.Errorf("expected 0 clients after unregister, got %d", hub.ClientCount())
	}
}

// TestHubUnregisterNonexistent tests unregistering a client that doesn't exist.
func TestHubUnregisterNonexistent(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, 256),
	}

	// Unregister without registering first - should not panic
	hub.unregister <- client
	time.Sleep(10 * time.Millisecond)

	if hub.ClientCount() != 0 {
		t.Errorf("expected 0 clients, got %d", hub.ClientCount())
	}
}

// TestHubBroadcast tests message broadcasting to registered clients.
func TestHubBroadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	// Create multiple mock clients
	client1 := &Client{hub: hub, send: make(chan []byte, 256)}
	client2 := &Client{hub: hub, send: make(chan []byte, 256)}
	client3 := &Client{hub: hub, send: make(chan []byte, 256)}

	// Register all clients
	hub.register <- client1
	hub.register <- client2
	hub.register <- client3
	time.Sleep(10 * time.Millisecond)

	if hub.ClientCount() != 3 {
		t.Fatalf("expected 3 clients, got %d", hub.ClientCount())
	}

	// Broadcast a message
	hub.Broadcast(WSMessage{Type: "test", Content: "hello"})
	time.Sleep(10 * time.Millisecond)

	// Check each client received the message
	for i, client := range []*Client{client1, client2, client3} {
		select {
		case msg := <-client.send:
			var wsMsg WSMessage
			if err := json.Unmarshal(msg, &wsMsg); err != nil {
				t.Errorf("client %d: failed to unmarshal message: %v", i+1, err)
				continue
			}
			if wsMsg.Type != "test" {
				t.Errorf("client %d: expected type 'test', got %q", i+1, wsMsg.Type)
			}
			if wsMsg.Content != "hello" {
				t.Errorf("client %d: expected content 'hello', got %q", i+1, wsMsg.Content)
			}
		default:
			t.Errorf("client %d did not receive message", i+1)
		}
	}
}

// TestHubBroadcastToNoClients tests broadcasting when no clients are connected.
func TestHubBroadcastToNoClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	// This should not panic or block
	hub.Broadcast(WSMessage{Type: "test"})
	time.Sleep(10 * time.Millisecond)
	// Success if no panic
}

// TestHubBroadcastDropsSlowClient tests that slow clients with full buffers get removed.
func TestHubBroadcastDropsSlowClient(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	// Create a client with a full send buffer
	slowClient := &Client{hub: hub, send: make(chan []byte, 1)} // Very small buffer
	slowClient.send <- []byte("blocking message")               // Fill the buffer

	hub.register <- slowClient
	time.Sleep(10 * time.Millisecond)

	if hub.ClientCount() != 1 {
		t.Fatalf("expected 1 client, got %d", hub.ClientCount())
	}

	// Broadcast - the slow client should be scheduled for removal
	hub.Broadcast(WSMessage{Type: "test"})
	time.Sleep(50 * time.Millisecond) // Give time for async unregister

	if hub.ClientCount() != 0 {
		t.Errorf("expected slow client to be removed, got %d clients", hub.ClientCount())
	}
}

// TestHubShutdown tests graceful hub shutdown.
func TestHubShutdown(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Register a client
	client := &Client{hub: hub, send: make(chan []byte, 256)}
	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	if hub.ClientCount() != 1 {
		t.Fatalf("expected 1 client, got %d", hub.ClientCount())
	}

	// Shutdown the hub
	hub.Shutdown()
	time.Sleep(50 * time.Millisecond)

	// After shutdown, client's send channel should be closed
	select {
	case _, ok := <-client.send:
		if ok {
			t.Error("expected send channel to be closed")
		}
	default:
		// Channel might be closed but empty - that's OK
	}
}

// TestNewClient tests client creation.
func TestNewClient(t *testing.T) {
	hub := NewHub()

	// We can't create a real websocket.Conn in unit tests easily,
	// so we test with nil (which is fine for the constructor)
	client := NewClient(hub, nil)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.hub != hub {
		t.Error("client hub is not set correctly")
	}
	if client.send == nil {
		t.Error("client send channel is nil")
	}
}

// TestWSMessageSerialization tests WSMessage JSON serialization.
func TestWSMessageSerialization(t *testing.T) {
	tests := []struct {
		name string
		msg  WSMessage
	}{
		{
			name: "tab_created",
			msg: WSMessage{
				Type: "tab_created",
				Tab: &Tab{
					ID:      "test-id",
					Title:   "Test Tab",
					Type:    TabTypeMarkdown,
					Content: "# Hello",
				},
			},
		},
		{
			name: "tab_deleted",
			msg: WSMessage{
				Type: "tab_deleted",
				ID:   "deleted-id",
			},
		},
		{
			name: "tab_activated",
			msg: WSMessage{
				Type: "tab_activated",
				ID:   "active-id",
			},
		},
		{
			name: "tabs_cleared",
			msg: WSMessage{
				Type: "tabs_cleared",
			},
		},
		{
			name: "with_content",
			msg: WSMessage{
				Type:    "custom",
				Content: "custom content",
			},
		},
		{
			name: "with_data",
			msg: WSMessage{
				Type: "data_message",
				Data: map[string]interface{}{
					"key": "value",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			data, err := json.Marshal(tt.msg)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			// Deserialize
			var result WSMessage
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			// Verify type is preserved
			if result.Type != tt.msg.Type {
				t.Errorf("expected type %q, got %q", tt.msg.Type, result.Type)
			}

			// Verify ID is preserved
			if result.ID != tt.msg.ID {
				t.Errorf("expected ID %q, got %q", tt.msg.ID, result.ID)
			}

			// Verify content is preserved
			if result.Content != tt.msg.Content {
				t.Errorf("expected content %q, got %q", tt.msg.Content, result.Content)
			}

			// Verify tab is preserved (if set)
			if tt.msg.Tab != nil {
				if result.Tab == nil {
					t.Error("expected Tab to be preserved")
				} else {
					if result.Tab.ID != tt.msg.Tab.ID {
						t.Errorf("expected Tab.ID %q, got %q", tt.msg.Tab.ID, result.Tab.ID)
					}
					if result.Tab.Title != tt.msg.Tab.Title {
						t.Errorf("expected Tab.Title %q, got %q", tt.msg.Tab.Title, result.Tab.Title)
					}
				}
			}
		})
	}
}

// TestWSMessageOmitEmpty tests that empty fields are omitted from JSON.
func TestWSMessageOmitEmpty(t *testing.T) {
	msg := WSMessage{Type: "test"}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	jsonStr := string(data)

	// Should not contain empty fields
	if strings.Contains(jsonStr, `"id"`) {
		t.Error("JSON should not contain empty id field")
	}
	if strings.Contains(jsonStr, `"tab"`) {
		t.Error("JSON should not contain nil tab field")
	}
	if strings.Contains(jsonStr, `"content"`) {
		t.Error("JSON should not contain empty content field")
	}

	// Should contain type
	if !strings.Contains(jsonStr, `"type":"test"`) {
		t.Error("JSON should contain type field")
	}
}

// TestHubBroadcastConcurrency tests concurrent broadcasts.
func TestHubBroadcastConcurrency(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	// Register multiple clients
	clients := make([]*Client, 10)
	for i := range clients {
		clients[i] = &Client{hub: hub, send: make(chan []byte, 256)}
		hub.register <- clients[i]
	}
	time.Sleep(20 * time.Millisecond)

	// Concurrently broadcast many messages
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			hub.Broadcast(WSMessage{Type: "test", Content: "message"})
		}(i)
	}
	wg.Wait()
	time.Sleep(50 * time.Millisecond)

	// All clients should still be connected
	if hub.ClientCount() != 10 {
		t.Errorf("expected 10 clients, got %d", hub.ClientCount())
	}
}

// TestHubRegisterConcurrency tests concurrent client registration.
func TestHubRegisterConcurrency(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &Client{hub: hub, send: make(chan []byte, 256)}
			hub.register <- client
		}()
	}
	wg.Wait()
	time.Sleep(50 * time.Millisecond)

	if hub.ClientCount() != 50 {
		t.Errorf("expected 50 clients, got %d", hub.ClientCount())
	}
}

// Integration tests that require a real HTTP server

// TestServeWS_Integration tests the WebSocket endpoint with a real HTTP server.
func TestServeWS_Integration(t *testing.T) {
	srv := NewServer()
	go srv.hub.Run()
	defer srv.hub.Shutdown()

	// Create a test server with the WebSocket handler
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWS(srv.hub, w, r, nil)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Connect a WebSocket client
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect WebSocket: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// Give time for registration
	time.Sleep(50 * time.Millisecond)

	if srv.hub.ClientCount() != 1 {
		t.Errorf("expected 1 client after connect, got %d", srv.hub.ClientCount())
	}
}

// TestServeWS_Broadcast tests that broadcasts reach WebSocket clients.
func TestServeWS_Broadcast(t *testing.T) {
	srv := NewServer()
	go srv.hub.Run()
	defer srv.hub.Shutdown()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWS(srv.hub, w, r, nil)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Connect a WebSocket client
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect WebSocket: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	time.Sleep(50 * time.Millisecond)

	// Broadcast a message
	srv.hub.Broadcast(WSMessage{
		Type: "tab_created",
		Tab:  &Tab{ID: "test", Title: "Test", Type: TabTypeMarkdown},
	})

	// Read the message from the client
	readCtx, readCancel := context.WithTimeout(ctx, 2*time.Second)
	defer readCancel()

	_, data, err := conn.Read(readCtx)
	if err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	var msg WSMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if msg.Type != "tab_created" {
		t.Errorf("expected type 'tab_created', got %q", msg.Type)
	}
	if msg.Tab == nil {
		t.Error("expected Tab to be set")
	} else if msg.Tab.ID != "test" {
		t.Errorf("expected Tab.ID 'test', got %q", msg.Tab.ID)
	}
}

// TestServeWS_ClientMessage tests that client messages are handled.
func TestServeWS_ClientMessage(t *testing.T) {
	srv := NewServer()
	go srv.hub.Run()
	defer srv.hub.Shutdown()

	// Create a tab first
	srv.state.CreateTab(&Tab{
		ID:    "test-tab",
		Title: "Test Tab",
		Type:  TabTypeMarkdown,
	})

	messageReceived := make(chan WSMessage, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWS(srv.hub, w, r, func(data []byte) {
			var msg WSMessage
			if json.Unmarshal(data, &msg) == nil {
				select {
				case messageReceived <- msg:
				default:
				}
			}
		})
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Connect a WebSocket client
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect WebSocket: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	time.Sleep(50 * time.Millisecond)

	// Send a message from the client
	clientMsg := WSMessage{Type: "activate_tab", ID: "test-tab"}
	msgData, _ := json.Marshal(clientMsg)
	if err := conn.Write(ctx, websocket.MessageText, msgData); err != nil {
		t.Fatalf("failed to write message: %v", err)
	}

	// Wait for message to be received
	select {
	case msg := <-messageReceived:
		if msg.Type != "activate_tab" {
			t.Errorf("expected type 'activate_tab', got %q", msg.Type)
		}
		if msg.ID != "test-tab" {
			t.Errorf("expected ID 'test-tab', got %q", msg.ID)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for message")
	}
}

// TestServeWS_MultipleClients tests multiple simultaneous WebSocket connections.
func TestServeWS_MultipleClients(t *testing.T) {
	srv := NewServer()
	go srv.hub.Run()
	defer srv.hub.Shutdown()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWS(srv.hub, w, r, nil)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Connect multiple WebSocket clients
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conns := make([]*websocket.Conn, 5)
	for i := range conns {
		conn, _, err := websocket.Dial(ctx, wsURL, nil)
		if err != nil {
			t.Fatalf("failed to connect client %d: %v", i, err)
		}
		conns[i] = conn
		defer conn.Close(websocket.StatusNormalClosure, "")
	}

	time.Sleep(100 * time.Millisecond)

	if srv.hub.ClientCount() != 5 {
		t.Errorf("expected 5 clients, got %d", srv.hub.ClientCount())
	}

	// Broadcast a message
	srv.hub.Broadcast(WSMessage{Type: "test", Content: "hello"})

	// All clients should receive the message
	for i, conn := range conns {
		readCtx, readCancel := context.WithTimeout(ctx, 2*time.Second)
		_, data, err := conn.Read(readCtx)
		readCancel()
		if err != nil {
			t.Errorf("client %d failed to read message: %v", i, err)
			continue
		}

		var msg WSMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			t.Errorf("client %d failed to unmarshal: %v", i, err)
			continue
		}

		if msg.Type != "test" {
			t.Errorf("client %d: expected type 'test', got %q", i, msg.Type)
		}
	}
}

// TestServeWS_ClientDisconnect tests that disconnected clients are removed.
func TestServeWS_ClientDisconnect(t *testing.T) {
	srv := NewServer()
	go srv.hub.Run()
	defer srv.hub.Shutdown()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWS(srv.hub, w, r, nil)
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Connect a WebSocket client
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if srv.hub.ClientCount() != 1 {
		t.Fatalf("expected 1 client, got %d", srv.hub.ClientCount())
	}

	// Disconnect the client
	conn.Close(websocket.StatusNormalClosure, "")
	time.Sleep(100 * time.Millisecond)

	if srv.hub.ClientCount() != 0 {
		t.Errorf("expected 0 clients after disconnect, got %d", srv.hub.ClientCount())
	}
}

// TestServerHandleWebSocket tests the server's WebSocket handler with message handling.
func TestServerHandleWebSocket(t *testing.T) {
	srv := NewServer()
	go srv.hub.Run()
	defer srv.hub.Shutdown()

	// Create tabs for testing
	srv.state.CreateTab(&Tab{
		ID:    "tab1",
		Title: "Tab 1",
		Type:  TabTypeMarkdown,
	})
	srv.state.CreateTab(&Tab{
		ID:    "tab2",
		Title: "Tab 2",
		Type:  TabTypeMarkdown,
	})

	// First tab should be active
	if srv.state.GetActive() != "tab1" {
		t.Fatalf("expected tab1 to be active initially, got %q", srv.state.GetActive())
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", srv.handleWebSocket)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	time.Sleep(50 * time.Millisecond)

	// Send activate_tab message
	activateMsg := WSMessage{Type: "activate_tab", ID: "tab2"}
	msgData, _ := json.Marshal(activateMsg)
	if err := conn.Write(ctx, websocket.MessageText, msgData); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	// Read the broadcast response
	readCtx, readCancel := context.WithTimeout(ctx, 2*time.Second)
	_, data, err := conn.Read(readCtx)
	readCancel()
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	var responseMsg WSMessage
	if err := json.Unmarshal(data, &responseMsg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if responseMsg.Type != "tab_activated" {
		t.Errorf("expected type 'tab_activated', got %q", responseMsg.Type)
	}
	if responseMsg.ID != "tab2" {
		t.Errorf("expected ID 'tab2', got %q", responseMsg.ID)
	}

	// Verify the tab was activated in state
	if srv.state.GetActive() != "tab2" {
		t.Errorf("expected tab2 to be active, got %q", srv.state.GetActive())
	}
}

// TestServerHandleWebSocket_CloseTab tests the close_tab message handling.
func TestServerHandleWebSocket_CloseTab(t *testing.T) {
	srv := NewServer()
	go srv.hub.Run()
	defer srv.hub.Shutdown()

	// Create a tab to delete
	srv.state.CreateTab(&Tab{
		ID:    "to-delete",
		Title: "To Delete",
		Type:  TabTypeMarkdown,
	})

	if srv.state.TabCount() != 1 {
		t.Fatalf("expected 1 tab, got %d", srv.state.TabCount())
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", srv.handleWebSocket)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	time.Sleep(50 * time.Millisecond)

	// Send close_tab message
	closeMsg := WSMessage{Type: "close_tab", ID: "to-delete"}
	msgData, _ := json.Marshal(closeMsg)
	if err := conn.Write(ctx, websocket.MessageText, msgData); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	// Read the broadcast response
	readCtx, readCancel := context.WithTimeout(ctx, 2*time.Second)
	_, data, err := conn.Read(readCtx)
	readCancel()
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	var responseMsg WSMessage
	if err := json.Unmarshal(data, &responseMsg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if responseMsg.Type != "tab_deleted" {
		t.Errorf("expected type 'tab_deleted', got %q", responseMsg.Type)
	}
	if responseMsg.ID != "to-delete" {
		t.Errorf("expected ID 'to-delete', got %q", responseMsg.ID)
	}

	// Verify the tab was deleted
	if srv.state.TabCount() != 0 {
		t.Errorf("expected 0 tabs after delete, got %d", srv.state.TabCount())
	}
}

// TestServerHandleWebSocket_InvalidMessage tests handling of invalid JSON messages.
func TestServerHandleWebSocket_InvalidMessage(t *testing.T) {
	srv := NewServer()
	go srv.hub.Run()
	defer srv.hub.Shutdown()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", srv.handleWebSocket)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	time.Sleep(50 * time.Millisecond)

	// Send invalid JSON - should not crash the server
	if err := conn.Write(ctx, websocket.MessageText, []byte("not json")); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Server should still be running and client should still be connected
	if srv.hub.ClientCount() != 1 {
		t.Errorf("expected client to still be connected, got %d clients", srv.hub.ClientCount())
	}
}

// TestServerHandleWebSocket_UnknownMessageType tests handling of unknown message types.
func TestServerHandleWebSocket_UnknownMessageType(t *testing.T) {
	srv := NewServer()
	go srv.hub.Run()
	defer srv.hub.Shutdown()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", srv.handleWebSocket)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	time.Sleep(50 * time.Millisecond)

	// Send unknown message type - should be silently ignored
	unknownMsg := WSMessage{Type: "unknown_type", ID: "some-id"}
	msgData, _ := json.Marshal(unknownMsg)
	if err := conn.Write(ctx, websocket.MessageText, msgData); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Server should still be running and client should still be connected
	if srv.hub.ClientCount() != 1 {
		t.Errorf("expected client to still be connected, got %d clients", srv.hub.ClientCount())
	}
}

// TestServerHandleWebSocket_ActivateNonexistentTab tests activating a tab that doesn't exist.
func TestServerHandleWebSocket_ActivateNonexistentTab(t *testing.T) {
	srv := NewServer()
	go srv.hub.Run()
	defer srv.hub.Shutdown()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", srv.handleWebSocket)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	time.Sleep(50 * time.Millisecond)

	// Send activate_tab for nonexistent tab - should not crash
	activateMsg := WSMessage{Type: "activate_tab", ID: "nonexistent"}
	msgData, _ := json.Marshal(activateMsg)
	if err := conn.Write(ctx, websocket.MessageText, msgData); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Server should still be running
	if srv.hub.ClientCount() != 1 {
		t.Errorf("expected client to still be connected, got %d clients", srv.hub.ClientCount())
	}
}

// TestServerHandleWebSocket_CloseNonexistentTab tests closing a tab that doesn't exist.
func TestServerHandleWebSocket_CloseNonexistentTab(t *testing.T) {
	srv := NewServer()
	go srv.hub.Run()
	defer srv.hub.Shutdown()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", srv.handleWebSocket)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	time.Sleep(50 * time.Millisecond)

	// Send close_tab for nonexistent tab - should not crash or broadcast
	closeMsg := WSMessage{Type: "close_tab", ID: "nonexistent"}
	msgData, _ := json.Marshal(closeMsg)
	if err := conn.Write(ctx, websocket.MessageText, msgData); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Server should still be running
	if srv.hub.ClientCount() != 1 {
		t.Errorf("expected client to still be connected, got %d clients", srv.hub.ClientCount())
	}
}

// TestServerHandleWebSocket_EmptyID tests messages with empty ID.
func TestServerHandleWebSocket_EmptyID(t *testing.T) {
	srv := NewServer()
	go srv.hub.Run()
	defer srv.hub.Shutdown()

	// Create a tab
	srv.state.CreateTab(&Tab{
		ID:    "existing",
		Title: "Existing",
		Type:  TabTypeMarkdown,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", srv.handleWebSocket)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	time.Sleep(50 * time.Millisecond)

	// Send activate_tab with empty ID - should be ignored
	activateMsg := WSMessage{Type: "activate_tab", ID: ""}
	msgData, _ := json.Marshal(activateMsg)
	if err := conn.Write(ctx, websocket.MessageText, msgData); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	// Send close_tab with empty ID - should be ignored
	closeMsg := WSMessage{Type: "close_tab", ID: ""}
	msgData, _ = json.Marshal(closeMsg)
	if err := conn.Write(ctx, websocket.MessageText, msgData); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Tab should still exist
	if srv.state.TabCount() != 1 {
		t.Errorf("expected tab to still exist, got %d tabs", srv.state.TabCount())
	}
}
