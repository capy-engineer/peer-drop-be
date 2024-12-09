package entity

import (
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

var Peers sync.Map

type PeerConnection struct {
	Conn       *websocket.Conn
	LastActive time.Time
}
