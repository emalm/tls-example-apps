package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type handler struct {
	verifyAppGuid      bool
	authorizedAppGuids []string
}

func NewHandler(appGuidList string) (*handler, error) {
	h := &handler{}
	if appGuidList == "" {
		return h, nil
	}

	err := json.Unmarshal([]byte(appGuidList), &h.authorizedAppGuids)
	if err != nil {
		return nil, err
	}

	h.verifyAppGuid = true
	return h, nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	info := extractClientInfo(req.TLS)
	err := h.authorize(info.appGuids)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(info.Name() + " not authorized: " + err.Error()))
		return
	}

	w.Write([]byte("authorized " + info.Name()))
}

type clientInfo struct {
	instance string
	appGuids []string
}

func (i clientInfo) Name() string {
	if len(i.appGuids) == 0 {
		return "client " + i.instance
	}

	return "instance " + i.instance + " of app " + strings.Join(i.appGuids, ", ")
}

func extractClientInfo(tlsInfo *tls.ConnectionState) clientInfo {
	leaf := tlsInfo.PeerCertificates[0]
	info := clientInfo{instance: leaf.Subject.CommonName}

	for _, ou := range leaf.Subject.OrganizationalUnit {
		if strings.HasPrefix(ou, "app:") {
			info.appGuids = append(info.appGuids, ou[4:])
		}
	}

	return info
}

var ErrNoMatch = errors.New("failed to find an authorized app guid")

func (h *handler) authorize(clientGuids []string) error {
	if !h.verifyAppGuid {
		return nil
	}

	for _, candidate := range clientGuids {
		for _, guid := range h.authorizedAppGuids {
			if candidate == guid {
				return nil
			}
		}
	}

	return ErrNoMatch
}
