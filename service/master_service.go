package service

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rhamerica/one/model"
)

// MasterService interface
type MasterService interface {
	Start()
	Stop()
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
	// /master/containers
	m.rs.Router().HandleFunc("/master/containers", func(w http.ResponseWriter, r *http.Request) { m.listContainers(w, r) }).Methods("GET")
	//router.HandleFunc("/master/", GetPeople).Methods("GET")
}

func (m *masterService) listContainers(w http.ResponseWriter, r *http.Request) []model.Container {
	//vars := mux.Vars(r)
	//filters []model.Filter
	def := map[string]string{
		"Name":           "string",
		"Image":          "string",
		"ContainerID":    "string",
		"HTTPPort":       "int",
		"DefinitionName": "string",
		"NodeName":       "string",
	}
	filters := model.RestFilters(def, r)
	containers = m.db.ListContainers()
	result := make([]model.Container, 0)
	for _, cont := range container {
		if model.FiltreMatch(cont, filters) {
			result = append(result, cont)
		}
	}
	return result
}
