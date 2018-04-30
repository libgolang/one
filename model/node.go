package model

// Node represents a server that hosts containers
type Node struct {
	Name    string
	Addr    string // of the form 10.10.10.1:8080
	Enabled bool
}
