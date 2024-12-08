package utils

import "github.com/google/uuid"

func IsValidPeerId(peerId string) bool {
	// Check if peerId is a valid UUID
	_, err := uuid.Parse(peerId)
	if err != nil {
		return false
	}
	return true
}
