package gskp

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"gopkg.in/tylerb/graceful.v1"

	"github.com/rs/xid"
	"github.com/utilitywarehouse/github-sshkey-provider/gskp/simplelog"
)

const (
	updateManagerInterval = time.Second

	defaultLongpollTimeoutDuration = 2 * time.Minute
)

var (
	serverInvalidParamTeam    = HTTPResponse{"error": "invalid team value"}
	serverInvalidParamInit    = HTTPResponse{"error": "invalid init value"}
	serverInvalidParamTimeout = HTTPResponse{"error": "invalid timeout value"}
	serverInvalidMethod       = HTTPResponse{"error": "invalid method"}
	serverUnexpectedError     = HTTPResponse{"error": "unexpected error occurred"}
	serverLongpollTimeout     = HTTPResponse{"error": "long polling has timed out"}
)

// HTTPResponse can be used to construct a response for an endpoint. It
// provides the Marshal function to serialise it.
type HTTPResponse map[string]interface{}

// Marshal will return the JSON encoded string of the HTTPResponse.
func (r HTTPResponse) Marshal() []byte {
	jsonText, _ := json.Marshal(r)

	return jsonText
}

// Server provides a proper but simple HTTP server that can handle graceful
// shutdowns and long polling requests.
type Server struct {
	cache                    *KeyCache
	mux                      *http.ServeMux
	server                   *graceful.Server
	updateManagerIsActive    bool
	updateManagerStopped     chan bool
	updateManagerQueue       map[string]map[string]chan bool
	updateManagerQueueMuxtex *sync.Mutex
}

// NewServer returns an instantiated Server which will use the provided
// KeyCache.
func NewServer(cache *KeyCache) (*Server, error) {
	mux := http.NewServeMux()

	ret := &Server{
		cache:                    cache,
		mux:                      mux,
		updateManagerStopped:     make(chan bool),
		updateManagerQueue:       map[string]map[string]chan bool{},
		updateManagerQueueMuxtex: &sync.Mutex{},
	}

	mux.HandleFunc("/status", ret.statusHandler)
	mux.HandleFunc("/keys", ret.keysHandler)

	return ret, nil
}

// Start will start listening for incoming connections.
func (s *Server) Start(listenAddress string, timeout time.Duration) error {
	s.server = &graceful.Server{
		Timeout: timeout,

		Server: &http.Server{
			Addr:    listenAddress,
			Handler: s.mux,
		},
	}

	simplelog.Infof("HTTP server listening on %s", listenAddress)

	s.updateManagerIsActive = true

	go s.updateManager()
	simplelog.Infof("update manager started")

	return s.server.ListenAndServe()
}

// Stop is a blocking operation that waits for the Server to close all
// connections and shutdown before returning.
func (s *Server) Stop(timeout time.Duration) {
	simplelog.Infof("HTTP server shutdown started with a timeout of %.0f seconds", timeout.Seconds())

	s.server.Stop(timeout)
	<-s.server.StopChan()
	simplelog.Infof("HTTP server shutdown complete")

	s.updateManagerIsActive = false
	<-s.updateManagerStopped
	simplelog.Infof("update manager stopped")
}

func (s *Server) updateManager() {
	for s.updateManagerIsActive {
		select {
		case team := <-s.cache.Updates:
			simplelog.Infof("received update message for team '%s', notifying clients", team)

			s.updateManagerQueueMuxtex.Lock()
			_, exists := s.updateManagerQueue[team]
			if exists {
				for _, notifier := range s.updateManagerQueue[team] {
					select {
					case notifier <- true:
					default:
					}
				}

				simplelog.Debugf("notified %d clients", len(s.updateManagerQueue[team]))
			} else {
				simplelog.Debugf("no clients are polling for team '%s'", team)
			}
			s.updateManagerQueueMuxtex.Unlock()
		case <-time.After(updateManagerInterval):
		}
	}

	s.updateManagerStopped <- true
}

func (s *Server) updateManagerGetNotifier(teamName string) (string, chan bool) {
	notifier := make(chan bool)
	notifierID := xid.New().String()

	s.updateManagerQueueMuxtex.Lock()
	if _, exists := s.updateManagerQueue[teamName]; !exists {
		s.updateManagerQueue[teamName] = map[string]chan bool{}
	}
	s.updateManagerQueue[teamName][notifierID] = notifier
	s.updateManagerQueueMuxtex.Unlock()

	return notifierID, notifier
}

func (s *Server) updateManagerRemoveNotifier(teamName string, notifierID string) {
	s.updateManagerQueueMuxtex.Lock()
	close(s.updateManagerQueue[teamName][notifierID])
	delete(s.updateManagerQueue[teamName], notifierID)
	s.updateManagerQueueMuxtex.Unlock()
}

func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.respond(w, http.StatusMethodNotAllowed, serverInvalidMethod)
		return
	}

	s.respond(
		w,
		http.StatusOK,
		HTTPResponse{
			"status":  "ok",
			"image":   os.Getenv("UW_IMAGE_NAME"),
			"git_sha": os.Getenv("UW_GIT_SHA"),
		},
	)
}

func (s *Server) keysHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.respond(w, http.StatusMethodNotAllowed, serverInvalidMethod)
		return
	}

	team := r.URL.Query().Get("team")
	init := r.URL.Query().Get("init")
	timeout := r.URL.Query().Get("timeout")

	if team == "" {
		s.respond(w, http.StatusBadRequest, serverInvalidParamTeam)
		return
	}

	if init != "" && init != "true" && init != "false" {
		s.respond(w, http.StatusBadRequest, serverInvalidParamInit)
		return
	}

	timeoutDuration := defaultLongpollTimeoutDuration
	if timeout != "" {
		t, err := strconv.ParseInt(timeout, 10, 64)
		if err != nil {
			s.respond(w, http.StatusBadRequest, serverInvalidParamTimeout)
			return
		}
		timeoutDuration = time.Duration(t) * time.Second
	}

	if init == "true" {
		if err := s.sendData(w, team); err != nil {
			simplelog.Errorf("error occurred when trying to get keys from cache: %v", err)
			s.respond(w, http.StatusInternalServerError, serverUnexpectedError)
			return
		}

		return
	}

	timeoutTimer := time.NewTimer(timeoutDuration)
	notifierID, notifier := s.updateManagerGetNotifier(team)
	defer s.updateManagerRemoveNotifier(team, notifierID)

	simplelog.Debugf("new longpoll connection '%s' from '%s' for team '%s'", notifierID, r.RemoteAddr, team)

	select {
	case <-notifier:
		timeoutTimer.Stop()

		if err := s.sendData(w, team); err != nil {
			simplelog.Errorf("error occurred when trying to get keys from cache: %v", err)
			s.respond(w, http.StatusInternalServerError, serverUnexpectedError)
			return
		}
	case <-timeoutTimer.C:
		simplelog.Debugf("timing out longpoll connection '%s' from '%s' for team '%s'", notifierID, r.RemoteAddr, team)
		s.respond(w, http.StatusOK, serverLongpollTimeout)
		return
	}
}

func (s *Server) respond(w http.ResponseWriter, code int, response HTTPResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response.Marshal())
}

func (s *Server) sendData(w http.ResponseWriter, teamName string) error {
	data, err := s.cache.Get(teamName)
	if err != nil {
		return err
	}

	simplelog.Debugf("responding to client with full data for team '%s'", teamName)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)

	return nil
}
