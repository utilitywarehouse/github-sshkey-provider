package gskp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gopkg.in/tylerb/graceful.v1"

	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

type httpResponse map[string]interface{}

func (r httpResponse) Marshal() string {
	jsonText, _ := json.Marshal(r)

	return string(jsonText)
}

// HTTPServer provides a proper but simple HTTP server that can handle graceful
// shutdowns.
type HTTPServer struct {
	mux    *http.ServeMux
	server *graceful.Server
}

// NewHTTPServer returns an instantiated HTTPServer.
func NewHTTPServer() *HTTPServer {
	ret := &HTTPServer{}

	mux := http.NewServeMux()
	mux.HandleFunc("/status", ret.endpointStatus)

	ret.mux = mux

	return ret
}

// Listen will start listening for incoming connections.
func (h *HTTPServer) Listen(address string) error {
	h.server = &graceful.Server{
		Timeout: 10 * time.Second,

		Server: &http.Server{
			Addr:    address,
			Handler: h.mux,
		},
	}

	h.server.ListenAndServe()

	return nil
}

// HandleGet registers a handler on the internal mux that will only accept GET
// requests.
func (h *HTTPServer) HandleGet(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	h.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, httpResponse{"error": "method not allowed"}.Marshal())
			return
		}

		simplelog.Debugf("Responding to %s request from %v", pattern, r.RemoteAddr)

		handler(w, r)
	})
}

// StopListening is a blocking operation that waits for the HTTPServer to close
// all connections and shutdown before returning.
func (h *HTTPServer) StopListening() {
	h.server.Stop(10 * time.Second)
	<-h.server.StopChan()
}

func (h *HTTPServer) endpointStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, httpResponse{"error": "method not allowed"}.Marshal())
		return
	}

	simplelog.Debugf("Responding to /status request from %v", r.RemoteAddr)

	fmt.Fprintf(w, httpResponse{
		"status": "ok",
	}.Marshal())
}
