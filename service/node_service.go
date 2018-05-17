package service

import (
	"time"

	"github.com/libgolang/log"
	"github.com/libgolang/one/clients"
	"github.com/libgolang/one/model"
)

// NodeService interface
type NodeService interface {
}

type nodeService struct {
	ticker       *time.Ticker
	masterClient clients.MasterClient
	docker       Docker
	nodeName     string
	nodeAddr     string
}

// NewNodeService NodeService constructor
func NewNodeService(masterAddr string, docker Docker, nodeName string, nodeAddr string) NodeService {
	ns := &nodeService{}
	ns.ticker = time.NewTicker(20 * time.Second)
	ns.masterClient = clients.NewMasterClient(masterAddr)
	ns.nodeName = nodeName
	ns.nodeAddr = nodeAddr
	ns.docker = docker
	ns.checkNode()
	go func() {
		for range ns.ticker.C {
			ns.checkNode()
		}
	}()
	return ns
}

func (n *nodeService) checkNode() {
	log.Info("checkNode()")

	node := model.Node{}
	node.Name = n.nodeName
	node.Addr = n.nodeAddr
	currentNfo := model.NodeInfo{}
	currentNfo.Containers = n.docker.ContainerList()
	currentNfo.Node = node

	infoFromMaster, err := n.masterClient.PingNodeInfo(currentNfo)
	if err != nil {
		log.Error("error contacting master: %s", err)
		return
	}

	currentMap := make(map[string]model.Container)
	serverMap := make(map[string]model.Container)

	for _, cont := range currentNfo.Containers {
		// remove dead container
		localCont := n.docker.ContainerGetByName(cont.Name)
		if localCont != nil && !localCont.Running {
			n.docker.ContainerRemoveByName(cont.Name)
			continue // continue
		}
		currentMap[cont.Name] = cont
	}

	for _, cont := range infoFromMaster.Containers {
		serverMap[cont.Name] = cont
	}
	log.Debug("%s", serverMap)
	log.Debug("%s", currentMap)

	// stop containers in currentMap that are not in serverMap
	for name := range currentMap {
		if _, ok := serverMap[name]; !ok {
			log.Info("Remove Container %s", name)
			n.docker.ContainerRemoveByName(name)
		}
	}

	// run containers not in currentMap that are in serverMap
	for name, cont := range serverMap {
		log.Info("Checking container %s:%s", name, cont.Name)
		if _, ok := currentMap[name]; !ok {
			// run
			log.Info("Running container %s", cont.Name)
			n.docker.ContainerRun(&cont)
		}
	}
}
