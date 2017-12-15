// Package handlers provides HTTP request handlers.
package handler

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"crowbar-stats/storage"
)

type route struct {
	pattern *regexp.Regexp
	handler http.Handler
}

type RegexpHandler struct {
	routes []*route
}

func (h *RegexpHandler) Handler(pattern *regexp.Regexp, handler http.Handler) {
	h.routes = append(h.routes, &route{pattern, handler})
}

func (h *RegexpHandler) HandleFunc(pattern *regexp.Regexp, handler func(http.ResponseWriter, *http.Request)) {
	h.routes = append(h.routes, &route{pattern, http.HandlerFunc(handler)})
}

func (h RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	for _, route := range h.routes {
		if route.pattern.MatchString(r.URL.Path) {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	// no pattern matched; send 404 response
	http.NotFound(w, r)
}

func urlcompile(url string) *regexp.Regexp {
	pattern, err := regexp.Compile(url)
	if err != nil {
		log.Fatalf("Cannot compile url, %v", url)
	}
	return pattern
}

// New returns an http handler for the url shortener.
// resource_history_url = "reports/nodes/#{@node.name}/runs/#{@run_id}"
// resource_history_url = "reports/nodes/#{node.name}/runs"
func New(storage storage.Service) http.Handler {
	mux := &RegexpHandler{}
	h := handler{storage}
	mux.HandleFunc(urlcompile("/reports/nodes/.*/runs/$"), responseHandler(h.newrun))
	mux.HandleFunc(urlcompile("/reports/nodes/.*/runs/.*"), responseHandler(h.rundata))
	return mux
}

type newrun struct {
	Uri         string `json:"uri"`
	SummaryOnly bool   `json:"summary_only"`
}

type response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"response"`
}

type handler struct {
	storage storage.Service
}

func responseHandler(h func(io.Writer, *http.Request) (int, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := h(w, r)
		w.WriteHeader(status)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			log.Printf("could not encode response to output: %v", err)
		}
	}
}

func (h handler) newrun(w io.Writer, r *http.Request) (int, error) {
	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, fmt.Errorf("method %s not allowed", r.Method)
	}

	var input map[string]string
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return http.StatusBadRequest, fmt.Errorf("Unable to decode JSON request body: %v, %v", err, r.Body)
	}

	action := strings.TrimSpace(input["action"])
	if action != "begin" {
		return http.StatusBadRequest, fmt.Errorf("payload is %v", input)
	}

	urlparts := strings.Split(r.URL.Path, "/")
	c, err := h.storage.Startrun(urlparts[3])
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("Could not store in database: %v", err)
	}

	err = json.NewEncoder(w).Encode(newrun{fmt.Sprintf("https://idontcare.com/%v", c), false})
	return http.StatusCreated, err
}

func (h handler) rundata(w io.Writer, r *http.Request) (int, error) {
	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, fmt.Errorf("method %s not allowed", r.Method)
	}

	reader, err := gzip.NewReader(r.Body)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("Unable to decode Gzip request body: %v", err)
	}

	defer reader.Close()
	id := strings.Split(r.URL.Path, "/")[5]

	payload, err := ioutil.ReadAll(reader)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("Could not read gzip content: %v", err)
	}
	_, err = h.storage.Stoprun(id, payload)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("Could not store in database: %v", err)
	}

	return http.StatusCreated, err
}
