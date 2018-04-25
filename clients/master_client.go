package clients

import "github.com/libgolang/one/model"

// MasterClient main interface
type MasterClient interface {
	ListContainersByNode(nodeName string) []model.Container
}

type masterClient struct {
}

// NewMasterClient constructor for MasterClient
func NewMasterClient() MasterClient {
	return &masterClient{}
}

func (m *masterClient) ListContainersByNode(nodeName string) []model.Container {
	return make([]model.Container, 0)
}
