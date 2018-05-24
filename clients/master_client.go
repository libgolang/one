package clients

import (
	"encoding/json"
	"fmt"

	"github.com/libgolang/log"

	"github.com/libgolang/one/model"
	"gopkg.in/resty.v1"
)

// MasterClient main interface
type MasterClient interface {
	//ListContainersByNode(nodeName string) []model.Container
	PingNodeInfo(nfo model.NodeInfo) (*model.NodeInfoResponse, error)
	GetDefinition(name string) (*model.Definition, error)
}

type masterClient struct {
	endPoint string
}

// NewMasterClient constructor for MasterClient
func NewMasterClient(masterAddr string) MasterClient {
	return &masterClient{fmt.Sprintf("http://%s", masterAddr)}
}

func (m *masterClient) ListContainersByNode(nodeName string) []model.Container {
	return make([]model.Container, 0)
}

func (m *masterClient) PingNodeInfo(nfo model.NodeInfo) (*model.NodeInfoResponse, error) {
	url := fmt.Sprintf("%s/master/nodeinfo", m.endPoint)
	jsonBody, err := json.Marshal(&nfo)
	if err != nil {
		return nil, err
	}
	log.Debug("POST %s\n%s\n", url, jsonBody)
	resp, err := resty.R().
		SetBody(jsonBody).
		SetHeader("Content-Type", "application/json").
		Post(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("Returned %d status code", resp.StatusCode())
	}

	nir := &model.NodeInfoResponse{}
	err = json.Unmarshal(resp.Body(), nir)
	return nir, err
}

func (m *masterClient) GetDefinition(name string) (*model.Definition, error) {
	resp, err := resty.R().
		SetPathParams(map[string]string{"name": name}).
		Get(fmt.Sprintf("%s/master/definitions/{name}", m.endPoint))

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("Returned %d status code", resp.StatusCode())
	}

	def := &model.Definition{}
	err = json.Unmarshal(resp.Body(), def)
	return def, err
}
