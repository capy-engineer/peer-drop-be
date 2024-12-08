package utils

import (
	"log"
	httpservice "peer-drop/internal/adapters/http"
	"time"
)

func RemoveInactivePeers() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		<-ticker.C
		httpservice.Peers.Range(func(key, value interface{}) bool {
			peerId := key.(string)
			peer := value.(httpservice.PeerConnection)
			if time.Since(peer.LastActive) > InactiveTimeout {
				err := peer.Conn.Close()
				if err != nil {
					return false
				}
				httpservice.Peers.Delete(peerId)
				log.Printf("Removed inactive peer: %s", peerId)
			}
			return true
		})
	}
}
