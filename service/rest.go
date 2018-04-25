package service

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/rhamerica/go-log"
	"golang.org/x/net/context"
)

// RestServer interface
type RestServer interface {
	Start()
	Stop()
	Router() *mux.Router
}

type restServer struct {
	srv    *http.Server
	router *mux.Router
}

// NewRestServer constructor of REST Server.
//   listenAddr of the form 127.0.0.1:8000
func NewRestServer(listenAddr string) MasterService {
	srv := &http.Server{}
	srv.Addr = listenAddr
	srv.WriteTimeout = 15 * time.Second
	srv.ReadTimeout = 15 * time.Second
	router := mux.NewRouter()
	return &restServer{srv: srv, router: router}
}

// Start method to start the service
func (m *restServer) Start() {
	//router.HandleFunc("/master/", GetPeople).Methods("GET")
	//router.HandleFunc("/people/{id}", GetPerson).Methods("GET")
	//router.HandleFunc("/people/{id}", CreatePerson).Methods("POST")
	//router.HandleFunc("/people/{id}", DeletePerson).Methods("DELETE")
	//log.Fatal(http.ListenAndServe(":8000", router))
	m.srv.Handler = m.router
	go func() {
		if err := m.srv.ListenAndServe(); err != nil {
			log.Warn("%s", err)
		}
	}()
}

func (m *restServer) Route() *mux.Router {
	return m.router
}

// Stop method to stop the service
func (m *restServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	m.srv.Shutdown(ctx)
}
