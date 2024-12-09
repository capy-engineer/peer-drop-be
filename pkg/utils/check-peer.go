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
			peerId := key.(string)
			peer := value.(entity.PeerConnection)
			if time.Since(peer.LastActive) > InactiveTimeout {
				err := peer.Conn.Close()
				if err != nil {
					return false
				}
				entity.Peers.Delete(peerId)
				log.Printf("Removed inactive peer: %s", peerId)
			}
			return true
		})
	}
}
