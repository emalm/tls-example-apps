package main

import (
	"encoding/json"
	"net/http"

	"github.com/emalm/tls-example-apps/discovery"
)

type discoveryHandler struct {
	ipAddress    string
	tlsPort      string
	instanceGuid string
}

func NewDiscoveryHandler(instanceGuid, ipAddress, tlsPort string) *discoveryHandler {
	return &discoveryHandler{
		instanceGuid: instanceGuid,
		ipAddress:    ipAddress,
		tlsPort:      tlsPort,
	}
}

func (h *discoveryHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	payload, err := json.Marshal(discovery.Location{
		InstanceGuid: h.instanceGuid,
		IPAddress:    h.ipAddress,
		TLSPort:      h.tlsPort,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error marshaling location data"))
		return
	}

	w.Write(payload)
}
