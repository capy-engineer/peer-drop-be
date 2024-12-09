package utils

import "github.com/google/uuid"

func IsValidPeerId(peerId string) error {
	// Check if peerId is a valid UUID
	_, err := uuid.Parse(peerId)
	if err != nil {
		return err
	}
	return nil
}
