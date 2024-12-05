package entity

// Device represents a device in the system.
type Device struct {
	Name    string `json:"name"`
	Service string `json:"service"`
	Address string `json:"address"`
}
