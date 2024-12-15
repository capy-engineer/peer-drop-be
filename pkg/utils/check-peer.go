package utils

import (
	"log"
	"peer-drop/internal/core/entity"
	"time"
)

func RemoveInactivePeers() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		<-ticker.C
		entity.Peers.Range(func(key, value interface{}) bool {
			peerId, ok := key.(string)
			if !ok {
				log.Println("Invalid key type in Peers map")
				return true
			}
			peer, ok := value.(entity.PeerConnection)
			if !ok {
				log.Println("Invalid value type in Peers map")
				return true // Continue the iteration
			}
			if time.Since(peer.LastActive) > InactiveTimeout {
				err := peer.Conn.Close()
				if err != nil {
					log.Printf("Failed to close connection for peer %s: %v", peerId, err)
					return true // Continue to the next peer
				}
				entity.Peers.Delete(peerId)
				log.Printf("Removed inactive peer: %s", peerId)
			}
			return true
		})
	}
}
