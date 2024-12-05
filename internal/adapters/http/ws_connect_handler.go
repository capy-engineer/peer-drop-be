package http

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var peers = make(map[string]*websocket.Conn)
var broadcast = make(chan []byte)

func SignalingHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	defer ws.Close()

	peerId := r.URL.Query().Get("peer_id")
	if peerId == "" {
		log.Println("Peer ID not found")
		return
	}
	peers[peerId] = ws
	log.Printf("Peer connected: %s", peerId)
}
