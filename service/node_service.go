package service

import (
	"time"

	log "github.com/rhamerica/go-log"
)

// NodeService interface
type NodeService interface {
}

type nodeService struct {
	listenIP   string
	listenPort int
	ticker     *time.Ticker
	masterEp   string
}

// NewNodeService NodeService constructor
func NewNodeService(listenIP string, listenPort int) NodeService {
	ns := &nodeService{}
	ns.listenIP = listenIP
	ns.listenPort = listenPort
	ns.ticker = time.NewTicker(20 * time.Second)
	go func() {
		for range ns.ticker.C {
			ns.checkNode()
		}
	}()
	return ns
}

func (n *nodeService) checkNode() {
	log.Info("checkNode()")

	n.runMissingContainers()
	// are all refs running?
	//	run missing ones

	n.stopZombiContainers()
	// are there any refs running that should not be?
	//	stop them
}

func (n *nodeService) runMissingContainers() {
	n.masterClient.getNodeDefinitions(n.nodeName)
}

func (n *nodeService) stopZombiContainers() {
}
