package checkapp

import "encoding/json"

// Info represents the health check information.
type Info struct {
	Status   string `json:"status"`
	Version  string `json:"version"`
	Hostname string `json:"hostname"`
}

// Encode implements web.Encoder interface.
func (i Info) Encode() ([]byte, string, error) {
	data, err := json.Marshal(i)
	return data, "application/json", err
}
