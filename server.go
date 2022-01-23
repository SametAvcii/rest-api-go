package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Coaster struct {
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
	ID           string `json:"id"`
	InPark       string `json:"inPark"`
	Height       int    `json:"height"`
}
type coastersHandlers struct {
	sync.Mutex
	store map[string]Coaster
}

func (h *coastersHandlers) coasters(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.get(w, r)
		return
	case "POST":
		h.post(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return
	}
}
func (h *coastersHandlers) get(w http.ResponseWriter, r *http.Request) {

	coasters := make([]Coaster, len(h.store))

	h.Lock()
	i := 0
	for _, coaster := range h.store {
		coasters[i] = coaster
		i++
	}
	h.Unlock()
	jsonBytes, err := json.Marshal(coasters)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}
func (h *coastersHandlers) getRandomCoaster(w http.ResponseWriter, r *http.Request) {
	ids := make([]string, len(h.store))
	h.Lock()
	i := 0
	for id := range h.store {

		ids[i] = id
		i++
	}

	defer h.Unlock()

	var target string
	if len(ids) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if len(ids) == 1 {
		target = ids[0]

	} else {
		rand.Seed(time.Now().UnixNano())
		target = ids[rand.Intn(len(ids))-1]
	}
	w.Header().Add("location", fmt.Sprintf("/coaster/%s", target))
	w.WriteHeader(http.StatusFound)
}
func (h *coastersHandlers) getCoaster(w http.ResponseWriter, r *http.Request) {

	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if parts[2] == "random" {
		h.getRandomCoaster(w, r)
	}

	h.Lock()
	coaster, ok := h.store[parts[2]]
	h.Unlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	jsonBytes, err := json.Marshal(coaster)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *coastersHandlers) post(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ct := r.Header.Get("content-type")
	if ct != "application/json" {

		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("need content-type 'application/json',but got'%s'", ct)))
		return

	}

	var coaster Coaster
	err = json.Unmarshal(bodyBytes, &coaster)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	coaster.ID = fmt.Sprintf("%d", time.Now().UnixNano())

	h.Lock()
	h.store[coaster.ID] = coaster
	defer h.Unlock()
}
func newCoasterHandlers() *coastersHandlers {
	return &coastersHandlers{
		store: map[string]Coaster{
			"id1": {
				Name:         "Anil44",
				Height:       250,
				ID:           "id1",
				InPark:       "Ataturk Parki",
				Manufacturer: "B+M",
			}, "id2": {

				Name:         "samet",
				Height:       30,
				ID:           "id2",
				InPark:       "dşlfd",
				Manufacturer: "ldfşdfs",
			},
		},
	}
}

type adminPortal struct {
	password string
}

func newAdminPortal() *adminPortal {
	password := os.Getenv("ADMİN_PASSWORD")
	if password == "" {
		panic("Required env var ADMİN PASSWORD not set")
	}
	return &adminPortal{password: password}
}
func (a adminPortal) handler(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok || user != "admin" || pass != a.password {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 - unauthorized"))
		return
	}
	w.Write([]byte("<html><h1>Super Secret admin Portal</h1></html>"))
}
func main() {
	admin := newAdminPortal()
	coastersHandlers := newCoasterHandlers()
	http.HandleFunc("/coasters", coastersHandlers.coasters)
	http.HandleFunc("/coasters/", coastersHandlers.getCoaster)
	http.HandleFunc("/admin", admin.handler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

//go run ADMIN_PASSWORD=anything server.go
