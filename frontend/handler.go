package main

import (
	"io/ioutil"
	"net/http"
)

type handler struct {
	client   *http.Client
	backends *Backends
}

func NewHandler(client *http.Client, backends *Backends) *handler {
	return &handler{
		client:   client,
		backends: backends,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	location, err := h.backends.Pick()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ip := location.IPAddress
	port := location.TLSPort

	resp, err := h.client.Get("https://" + ip + ":" + port)
	if err != nil {
		h.backends.Remove(location)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error making request to " + ip + ":" + port + ": " + err.Error()))
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error reading response: " + err.Error()))
		return
	}

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("backend " + location.Name() + " failed to authorize frontend: "))
		w.Write(body)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success from backend " + location.Name() + ": "))
	w.Write(body)
}
