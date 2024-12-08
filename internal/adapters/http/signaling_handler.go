package httpservice

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections only from trusted origins
		//origin := r.Header.Get("Origin")
		//allowedOrigins := map[string]bool{
		//	"https://trusted-domain.com": true,
		//	"https://another-trusted.com": true,
		//}
		//return allowedOrigins[origin]
		return true
	},
}

var peers sync.Map

const inactiveTimeout = 10 * time.Minute

type PeerConnection struct {
	Conn       *websocket.Conn
	LastActive time.Time
}

func SignalingHandler(c echo.Context) error {
	conn, err := upgrade.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to upgrade to WebSocket")
	}

	peerId := c.QueryParam("peerId")
	if peerId == "" {
		peerId, err := uuid.NewV7()
		if err != nil {
			log.Printf("Error generating UUID: %v", err)
			err := conn.Close()
			if err != nil {
				return err
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate peerId")
		}

		err = conn.WriteMessage(websocket.TextMessage, []byte(peerId.String()))
		if err != nil {
			log.Printf("Error sending UUID to client: %v", err)
			err := conn.Close()
			if err != nil {
				return err
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to send peerId")
		}
	}

	// Close old connection if it exists
	if v, ok := peers.Load(peerId); ok {
		oldPeer := v.(PeerConnection)
		err := oldPeer.Conn.Close()
		if err != nil {
			return err
		}
		log.Printf("Closed old connection for peerId: %s", peerId)
	}

	peers.Store(peerId, PeerConnection{Conn: conn, LastActive: time.Now()})
	log.Printf("Stored new connection for peerId: %s", peerId)

	defer func() {
		if v, ok := peers.Load(peerId); ok {
			peer := v.(PeerConnection)
			err := peer.Conn.Close()
			if err != nil {
				return
			}
			peers.Delete(peerId)
			log.Printf("Removed connection for peerId: %s", peerId)
		}
	}()

	// Heartbeat to update LastActive
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			<-ticker.C
			if _, ok := peers.Load(peerId); !ok {
				return // Stop heartbeat if peer is removed
			}
			peers.Store(peerId, PeerConnection{Conn: conn, LastActive: time.Now()})
		}
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from %s: %v", peerId, err)
			break
		}

		// Parse incoming message
		var payload map[string]interface{}
		if err := json.Unmarshal(msg, &payload); err != nil {
			log.Printf("Invalid message format from %s: %v", peerId, err)
			err := conn.WriteMessage(websocket.TextMessage, []byte("Error: Invalid message format"))
			if err != nil {
				return err
			}
			continue
		}

		// Validate and forward message
		targetId, ok := payload["targetId"].(string)
		if !ok || targetId == "" {
			log.Printf("Missing or invalid targetId in message from %s", peerId)
			err := conn.WriteMessage(websocket.TextMessage, []byte("Error: Missing or invalid targetId"))
			if err != nil {
				return err
			}
			continue
		}

		if targetPeer, ok := peers.Load(targetId); ok {
			targetConn := targetPeer.(PeerConnection).Conn
			if err := targetConn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("Error forwarding message to %s: %v", targetId, err)
			}
		} else {
			log.Printf("Target peer not found: %s", targetId)
			err := conn.WriteMessage(websocket.TextMessage, []byte("Error: Target peer not found"))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func RemoveInactivePeers() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		<-ticker.C
		peers.Range(func(key, value interface{}) bool {
			peerId := key.(string)
			peer := value.(PeerConnection)
			if time.Since(peer.LastActive) > inactiveTimeout {
				err := peer.Conn.Close()
				if err != nil {
					return false
				}
				peers.Delete(peerId)
				log.Printf("Removed inactive peer: %s", peerId)
			}
			return true
		})
	}
}
