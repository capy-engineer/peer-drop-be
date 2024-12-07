package httpservice

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"sync"
)

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var peers sync.Map

func SignalingHandler(c echo.Context) error {

	conn, err := upgrade.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to upgrade to WebSocket")
	}
	defer func() {
		peerId := c.QueryParam("peerId")
		if peerId != "" {
			peers.Delete(peerId)
			log.Printf("Peer disconnected: %s", peerId)
		}
		conn.Close()
	}()

	peerId := c.QueryParam("peerId")
	if peerId == "" {
		uid, err := uuid.NewV7()
		if err != nil {
			log.Printf("Error generating UUID: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate UUID")
		}
		peerId = uid.String()
		log.Printf("Generated new peerId: %s", peerId)
	}

	// Store the connection
	peers.Store(peerId, conn)
	log.Printf("Peer connected: %s", peerId)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from %s: %v", peerId, err)
			break
		}

		// Parse incoming message
		var payload map[string]interface{}
		if err := json.Unmarshal(msg, &payload); err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}

		targetId, ok := payload["targetId"].(string)
		if !ok || targetId == "" {
			log.Printf("Missing or invalid targetId in message from %s", peerId)
			continue
		}

		if targetConn, ok := peers.Load(targetId); ok {
			if err := targetConn.(*websocket.Conn).WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("Error forwarding message to %s: %v", targetId, err)
			}
		} else {
			log.Printf("Target peer not found: %s", targetId)
		}
	}

	return nil
}
