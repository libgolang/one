package service

import (
	"fmt"
	"math/rand"
	"net/http"

	"time"

	"github.com/libgolang/log"
	"github.com/libgolang/one/model"
	"github.com/libgolang/one/utils"
)

const (
	masterTick = time.Second * 10
)

// MasterService interface
type MasterService interface {
}

type masterService struct {
	db Db
	rs RestServer
}

// NewMasterService constructor of Master REST API.
func NewMasterService(rs RestServer, db Db) MasterService {
	master := &masterService{rs: rs, db: db}
	master.init()
	return master
}

func (m *masterService) init() {
	// api
	m.rs.HandleFunc("/master/containers", func(w http.ResponseWriter, r *http.Request) RestResponse { return m.listContainers(w, r) }).Methods("GET")
	m.rs.HandleFunc("/master/nodes", func(w http.ResponseWriter, r *http.Request) RestResponse { return m.listNodes(w, r) }).Methods("GET")
	m.rs.HandleFunc("/master/nodes", func(w http.ResponseWriter, r *http.Request) RestResponse { return m.upsertNode(w, r) }).Methods("POST")

	// process definitions
	timer := time.NewTicker(masterTick)
	go func() {
		for range timer.C {
			m.allocateContainers()
			log.Info("Tick")
		}
	}()
}

func (m *masterService) listNodes(w http.ResponseWriter, r *http.Request) RestResponse {
	def := map[string]string{
		"Name":    "string",
		"Enabled": "bool",
	}
	list := m.db.ListNodes()
	utils.RestFilterReduce(def, r, &list)
	return (&JSONResponse{}).SetBody(list)
}

func (m *masterService) upsertNode(w http.ResponseWriter, r *http.Request) RestResponse {
	// TODO
	panic("not implemented")
}

// This looks at the definitions and containers and makes sure that
// all the container records are distributed evenly among all nodes
func (m *masterService) allocateContainers() {
	//nodeMap := m.db.ListNodes()
	defMap := m.db.ListDefinitions()
	contMap := m.db.ListContainers()

	//
	// definition -> container map
	//
	defContMapList := make(map[string][]*model.Container)
	for _, cont := range contMap {
		conts := defContMapList[cont.DefinitionName]
		conts = append(conts, cont)
		defContMapList[cont.DefinitionName] = conts
	}

	//
	// node -> container map
	//
	nodeContMap := make(map[string][]*model.Container)
	for _, cont := range contMap {
		conts := nodeContMap[cont.NodeName]
		conts = append(conts, cont)
		nodeContMap[cont.NodeName] = conts
	}

	//
	// todo
	//
	for k, def := range defMap {
		conts := defContMapList[k]
		n := len(conts)
		if def.Count < n {
			// deallocate some containers for definition
			diff := n - def.Count
			log.Info("Adjusting container count (%d delta)", diff)
			for i := 0; i < diff; i++ {
				idx := rand.Intn(len(conts))
				cont := conts[idx]
				conts = append(conts[:idx], conts[idx+1:]...)
				log.Info("Deleting container id %s/%s", cont.ContainerID, cont.Name)
				m.db.DeleteContainer(cont.ContainerID)
			}
		} else if def.Count > n {
			// allocate more containers for definition
			diff := def.Count - n
			log.Info("Adjusting container count (%d delta)", diff)
			for i := 0; i < diff; i++ {
				c := &model.Container{}
				c.Name = fmt.Sprintf("%s-%d", def.Name, m.db.NextAutoIncrement("inc.container", def.Name))
				c.DefinitionName = def.Name

				//
				// find node with least numbers of containers
				//
				currentN := 999999999
				var currentNodeName string
				for nodeName, contSlice := range nodeContMap {
					n := len(contSlice)
					if currentN > n {
						currentNodeName = nodeName
						currentN = n
					}

				}
				c.NodeName = currentNodeName
				if currentNodeName == "" {
					log.Warn("Not able to create node %s...no nodes available!", c.Name)
					continue
				}

				//
				c.Image = def.Image
				c.Running = false
				c.HTTPPort = m.db.NextAutoIncrement("http.port", "http.port")
				log.Info("Creating container id %s/%s", c.ContainerID, c.Name)
				if err := m.db.SaveContainer(c); err != nil {
					log.Error("Error saving container %s", c.Name)
				} else {
					nodeContMap[c.NodeName] = append(nodeContMap[c.NodeName], c)
				}
			}
		}
	}
}

func (m *masterService) listContainers(w http.ResponseWriter, r *http.Request) RestResponse {
	def := map[string]string{
		"Name":           "string",
		"Image":          "string",
		"ContainerID":    "string",
		"HTTPPort":       "int",
		"DefinitionName": "string",
		"NodeName":       "string",
	}
	containers := m.db.ListContainers()
	utils.RestFilterReduce(def, r, &containers)
	resp := (&JSONResponse{}).SetBody(containers)
	return resp
}
