package main

import "fmt"

// MasterClient main interface
type MasterClient interface {
 GetContainersByNode(nodeName) []mode.Container 
}

type masterClient struct {
}

// NewMasterClient constructor for MasterClient
func NewMasterClient() MasterClient {
	return &masterClient{
	}
}


func (m *masterClient) GetContainersByNode(nodeName) []mode.Container {
	m.db.
}
