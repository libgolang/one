package model

import "time"

// Node represents a server that hosts containers
type Node struct {
	Name        string    `json:"name"`
	Addr        string    `json:"addr"` // of the form 10.10.10.1:8080
	Enabled     bool      `json:"enabled"`
	LastUpdated time.Time `json:"lastUpdated"`
}
