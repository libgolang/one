package service

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/libgolang/log"
	"golang.org/x/net/context"
)

// RestServer interface
type RestServer interface {
	Start()
	StartAndBlock()
	Stop()
	HandleFunc(path string, f func(w http.ResponseWriter, r *http.Request) RestResponse) *mux.Route
}

type restServer struct {
	srv      *http.Server
	router   *mux.Router
	certFile string
	keyFile  string
}

// RestResponse response interface
type RestResponse interface {
	ContentType() string
	Body() interface{}
	Status() int
	Headers() map[string]string
}

// NewRestServer constructor of REST Server.
//   listenAddr of the form 127.0.0.1:8000
func NewRestServer(listenAddr, certFile, keyFile string) RestServer {
	srv := &http.Server{}
	srv.Addr = listenAddr
	srv.WriteTimeout = 15 * time.Second
	srv.ReadTimeout = 15 * time.Second
	router := mux.NewRouter()
	return &restServer{
		srv:      srv,
		router:   router,
		certFile: certFile,
		keyFile:  keyFile,
	}
}

// Start method to start the service
func (m *restServer) Start() {
	//router.HandleFunc("/master/", GetPeople).Methods("GET")
	//router.HandleFunc("/people/{id}", GetPerson).Methods("GET")
	//router.HandleFunc("/people/{id}", CreatePerson).Methods("POST")
	//router.HandleFunc("/people/{id}", DeletePerson).Methods("DELETE")
	//log.Fatal(http.ListenAndServe(":8000", router))
	go func() {
		m.StartAndBlock()
	}()
}

func (m *restServer) StartAndBlock() {
	m.srv.Handler = m.router
	var err error
	if m.certFile == "" || m.keyFile == "" {
		err = m.srv.ListenAndServe()
	} else {
		err = m.srv.ListenAndServeTLS(m.certFile, m.keyFile)
	}
	if err != nil {
		log.Warn("%s", err)
	}
}

func (m *restServer) Router() *mux.Router {
	return m.router
}
func (m *restServer) HandleFunc(path string, f func(w http.ResponseWriter, r *http.Request) RestResponse) *mux.Route {
	//m.rs.Router().HandleFunc("/master/containers", func(w http.ResponseWriter, r *http.Request) { m.listContainers(w, r) }).Methods("GET")
	return m.router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		ret := f(w, r)
		if ret == nil {
			return
		}
		bytes, err := json.Marshal(ret.Body())
		if err != nil {
			log.Error("%s", err)
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(`{"error": "Internal Server Error"}`))
			if err != nil {
				log.Error("%s", err)
			}
			return
		}

		w.Header().Set("Content-Type", ret.ContentType())
		if _, err := w.Write(bytes); err != nil {
			log.Error("%s", err)
		}

		for k, v := range ret.Headers() {
			w.Header().Set(k, v)
		}
	})
}

// Stop method to stop the service
func (m *restServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := m.srv.Shutdown(ctx); err != nil {
		log.Error("error shutting down rest server: %s", err)
	}
}

// JSONResponse RestResponse implementation
type JSONResponse struct {
	headers     map[string]string
	contentType string
	body        interface{}
	status      int
}

// Headers returns the headers
func (j *JSONResponse) Headers() map[string]string {
	return j.headers
}

// Status http status
func (j *JSONResponse) Status() int {
	var s int
	if j.status == 0 {
		s = 200
	} else {
		s = j.status
	}
	return s
}

// Body the body object to return
func (j *JSONResponse) Body() interface{} {
	return j.body
}

// ContentType  content-type
func (j *JSONResponse) ContentType() string {
	if j.contentType != "" {
		return j.contentType
	}
	return "application/json"
}

// SetStatus http status
func (j *JSONResponse) SetStatus(s int) *JSONResponse {
	j.status = s
	return j
}

// SetBody the body object to return
func (j *JSONResponse) SetBody(b interface{}) *JSONResponse {
	j.body = b
	return j
}

// SetContentType  content-type
func (j *JSONResponse) SetContentType(c string) *JSONResponse {
	j.contentType = c
	return j
}
