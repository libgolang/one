package service

import (
	"github.com/rhamerica/one/model"

	"github.com/rhamerica/go-log"
)

// Cycle interface for cycle
type Cycle interface {
	Start(*model.Definition)
	Stop(*model.Definition)
}

type cycle struct {
	hostIP string
	db     Db
	p      Proxy
	docker Docker
}

// NewLifecycle constructor
func NewLifecycle(hostIP string, db Db, p Proxy, docker Docker) Cycle {
	return &cycle{
		hostIP: hostIP,
		db:     db,
		p:      p,
		docker: docker,
	}
}

func (c *cycle) Stop(def *model.Definition) {
	log.Info("Stopping container definition '%s'", def.Name)

	// Execute
	if c.docker.IsRunningByDefName(def.Name) {
		c.docker.ContainerStopByDefName(def.Name)
	}

	if c.docker.ContainerExists(def.Name) {
		c.docker.ContainerRemoveByDefName(def.Name)
	}

	//
	c.p.RemoveDockerProxy(def.Name)
}

func (c *cycle) Start(def *model.Definition) {
	log.Info("Starting container definition '%s'", def.Name)

	//
	cont := c.docker.ContainerRun(def)

	//
	log.Debug("container http port: %d", cont.HTTPPort)
	if cont.HTTPPort > 0 {
		c.p.AddDockerProxySsl(def.Name, c.hostIP, cont.HTTPPort)
	}
}
