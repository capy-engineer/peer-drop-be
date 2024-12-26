package httpservice

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"peer-drop/internal/core/entity"
	"peer-drop/pkg/utils"
	"time"
)

func ConnectHandler(c echo.Context) error {
	conn, err := upgrade.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to upgrade to WebSocket")
	}

	var peerId string
	var peerTarget string

	var peerIdClient = c.QueryParam("peerId")
	if peerIdClient != "" {
		err := utils.IsValidPeerId(peerIdClient)
		if err != nil {
			log.Printf("Invalid peerId: %v", err)
			err := conn.Close()
			if err != nil {
				return err
			}
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid peerId")
		}
		peerTarget = peerIdClient
	}
	uid, err := uuid.NewV7()
	if err != nil {
		log.Printf("Error generating UUID: %v", err)
		err := conn.Close()
		if err != nil {
			return err
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate peerId")
	}
	peerId = uid.String()

	// Close old connection if it exists
	if v, ok := entity.Peers.Load(peerId); ok {
		oldPeer := v.(entity.PeerConnection)
		err := oldPeer.Conn.Close()
		if err != nil {
			return err
		}
		log.Printf("Closed old connection for peerId: %s", peerId)
	}

	// Add new connection to map
	entity.Peers.Store(peerId, entity.PeerConnection{Conn: conn, LastActive: time.Now()})
	log.Printf("Stored new connection for peerId: %s", peerId)

	defer func() {
		if v, ok := entity.Peers.Load(peerId); ok {
			peer := v.(entity.PeerConnection)
			err := peer.Conn.Close()
			if err != nil {
				log.Printf("Error closing connection for peerId: %s, %v", peerId, err)
			}
			entity.Peers.Delete(peerId)
			log.Printf("Removed connection for peerId: %s", peerId)
		}
	}()

	// If peerTarget is specified, connect to peer
	if peerTarget != "" {
		if targetPeer, ok := entity.Peers.Load(peerTarget); ok {
			targetConn := targetPeer.(entity.PeerConnection).Conn
			err := targetConn.WriteJSON(map[string]string{
				"type":   "connect",
				"peerId": peerId,
			})
			if err != nil {
				log.Printf("Error notifying target peer %s: %v", peerTarget, err)
			}
		} else {

			log.Printf("Target peer %s not found", peerTarget)
			if v, ok := entity.Peers.Load(peerId); ok {
				peer := v.(entity.PeerConnection)
				err := peer.Conn.Close()
				if err != nil {
					log.Printf("Error closing connection for peerId: %s, %v", peerId, err)
				}
				entity.Peers.Delete(peerId)
				log.Printf("Removed connection for peerId: %s", peerId)
			}
		}
	}

	// Heartbeat to update LastActive
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			<-ticker.C
			if _, ok := entity.Peers.Load(peerId); !ok {
				return // Stop heartbeat if peer is removed
			}
			entity.Peers.Store(peerId, entity.PeerConnection{Conn: conn, LastActive: time.Now()})
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

		payload["senderId"] = peerId
		fmt.Println("user B", payload)

		if targetPeer, ok := entity.Peers.Load(targetId); ok {
			targetConn := targetPeer.(entity.PeerConnection).Conn
			if err := targetConn.WriteJSON(payload); err != nil {
				log.Printf("Error forwarding message to %s: %v", targetId, err)
			}
		} else {
			err := conn.WriteJSON(map[string]string{
				"type":  "error",
				"error": "Target peer not found",
				"info":  fmt.Sprintf("TargetId %s does not exist or is offline", targetId),
			})
			if err != nil {
				log.Printf("Error sending error message to %s: %v", peerId, err)
			}
		}
	}

	return nil
}
