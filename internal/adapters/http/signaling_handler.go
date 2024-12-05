package httpservice

import (
	"encoding/json"
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
		return err
	}
	defer func() {
		peerId := c.QueryParam("peer_id")
		if peerId != "" {
			peers.Delete(peerId)
			log.Printf("Peer disconnected: %s", peerId)
		}
		conn.Close()
	}()

	peerId := c.QueryParam("peer_id")
	if peerId == "" {
		log.Println("Missing peer ID")
		return echo.NewHTTPError(http.StatusBadRequest, "Missing peer ID")
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
		
		targetId := payload["target_id"].(string)
		if targetConn, ok := peers.Load(targetId); ok {
			// Forward message to target peer
			targetConn.(*websocket.Conn).WriteMessage(websocket.TextMessage, msg)
		} else {
			log.Printf("Target peer not found: %s", targetId)
		}
	}

	return nil
}
