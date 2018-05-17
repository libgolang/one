package service

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"

	"encoding/json"
	"time"

	"github.com/gorilla/mux"
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
	m.rs.HandleFunc("/master/nodeinfo", func(w http.ResponseWriter, r *http.Request) RestResponse { return m.pingNodeInfo(w, r) }).Methods("POST")
	m.rs.HandleFunc("/master/definitions/{name}", func(w http.ResponseWriter, r *http.Request) RestResponse { return m.getDefinition(w, r) }).Methods("GET")

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

func (m *masterService) pingNodeInfo(w http.ResponseWriter, r *http.Request) RestResponse {
	log.Info("POST %s : pingNodeInfo()", r.URL)
	resp := &JSONResponse{}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("error reading body: %s", err)
		return resp.SetStatus(400).SetBody(`{"error":"Unable to read request"}`)
	}

	nfo := &model.NodeInfo{}
	err = json.Unmarshal(b, nfo)
	if err != nil {
		log.Error("error decoding json: %s", err)
		return resp.SetStatus(400).SetBody(`{"error":"Unable to parse request"}`)
	}

	if nfo.Node.Name == "" || nfo.Node.Addr == "" {
		return resp.SetStatus(400).SetBody(`{"error":"Invalid Node Information"}`)
	}

	// Register / Update Nodes
	m.db.Trx(func(db Db) {
		log.Debug("#############################################################")
		node, err := db.GetNode(nfo.Node.Name)
		if err != nil {
			// create
			node = &model.Node{}
			node.Name = nfo.Node.Name
			node.Enabled = true
		}
		node.LastUpdated = time.Now()
		node.Addr = nfo.Node.Addr

		err = db.SaveNode(node)
		if err != nil {
			log.Error("%s", err)
		}
		log.Debug("#############################################################")
	})

	// Make sure containers match
	officialContainers := m.db.ListContainers()
	for _, cont := range nfo.Containers {
		_, ok := officialContainers[cont.Name]
		if !ok {
			log.Warn("Node %s has an unknown container %s", nfo.Node.Name, cont.Name)
		}
	}

	// Respond with the list of containers in file
	node := nfo.Node
	containers := make([]model.Container, 0)
	for _, cont := range m.db.ListContainers() {
		if node.Name == cont.NodeName {
			containers = append(containers, *cont)
		}
	}
	return resp.SetBody(&model.NodeInfoResponse{Containers: containers})
}

// This looks at the definitions and containers and makes sure that
// all the container records are distributed evenly among all nodes
func (m *masterService) allocateContainers() {
	defMap := m.db.ListDefinitions()
	contMap := m.db.ListContainers()
	nodeMap := m.db.ListNodes()

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
	for nodeName := range nodeMap {
		conts := make([]*model.Container, 0)
		for _, cont := range contMap {
			if cont.NodeName == nodeName {
				conts = append(conts, cont)
			}
		}
		nodeContMap[nodeName] = conts
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
				log.Debug("nodeContMap: %s", nodeContMap)
				for nodeName, contSlice := range nodeContMap {
					n := len(contSlice)
					log.Debug("checking node for number of containers (%d) less than %d", n, currentN)
					if currentN > n {
						currentNodeName = nodeName
						currentN = n
					}

				}
				if currentNodeName == "" {
					log.Warn("Not able to create node %s...no nodes available!", c.Name)
					continue
				}
				c.NodeName = currentNodeName

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

func (m *masterService) getDefinition(w http.ResponseWriter, r *http.Request) RestResponse {
	resp := &JSONResponse{}
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return resp.SetStatus(400).SetBody(`{"error":"Name is required"}`)
	}

	defPtr, err := m.db.GetDefinition(name)
	if err != nil {
		log.Error("Definition not found: %s", err)
		return resp.SetStatus(404).SetBody(`{"error":"definition not found"}`)
	}

	return resp.SetBody(defPtr)
}
