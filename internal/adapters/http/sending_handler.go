package httpservice

import (
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"peer-drop/internal/core/entity"
	"peer-drop/pkg/utils"
)

func SendingHandler(c echo.Context) error {
	conn, err := upgrade.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to upgrade to WebSocket")
	}

	var targetId string
	targetId = c.QueryParam("targetId")
	if targetId != "" {
		err := utils.IsValidPeerId(targetId)
		if err != nil {
			log.Printf("Invalid targetId: %v", err)
			err := conn.Close()
			if err != nil {
				return err
			}
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid targetId")
		}

		if targetPeer, ok := entity.Peers.Load(targetId); ok {
			targetConn := targetPeer.(entity.PeerConnection).Conn
			err := targetConn.WriteJSON(map[string]string{
				"type":  "offer",
				"offer": targetId,
			})
			if err != nil {
				log.Printf("Error notifying target peer %s: %v", targetId, err)
			}
		} else {

			log.Printf("Target peer %s not found", targetId)
		}
	}

	return nil

}
