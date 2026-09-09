package ws_test

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pos-backend/internal/adapters/handlers/ws"
)

func TestWebSocketHub_Broadcast(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hub := ws.NewHub()
	go hub.Run()

	r := gin.New()
	r.GET("/ws", hub.ServeWS)

	server := httptest.NewServer(r)
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	// Connect WebSocket client
	wsConn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	defer wsConn.Close()

	// Give time for hub to register client
	time.Sleep(50 * time.Millisecond)

	// Broadcast stock update event
	hub.BroadcastStockUpdate("SKU-COFFEE-100", 15)

	// Read message from WebSocket client
	wsConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, p, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read websocket message: %v", err)
	}

	var event ws.StockUpdateEvent
	if err := json.Unmarshal(p, &event); err != nil {
		t.Fatalf("failed to unmarshal JSON event: %v", err)
	}

	if event.Type != "stock_update" {
		t.Errorf("expected type stock_update, got %s", event.Type)
	}
	if event.SKU != "SKU-COFFEE-100" {
		t.Errorf("expected SKU SKU-COFFEE-100, got %s", event.SKU)
	}
	if event.NewStock != 15 {
		t.Errorf("expected new stock 15, got %d", event.NewStock)
	}
}
