package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type handler struct {
	authorizedAppGuids []string
}

func NewHandler(appGuidList string) (*handler, error) {
	h := &handler{}
	if appGuidList == "" {
		return h, nil
	}

	appGuids := []string{}
	err := json.Unmarshal([]byte(appGuidList), &appGuids)
	if err != nil {
		return nil, err
	}

	h.authorizedAppGuids = appGuids
	return h, nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	appGuid, instanceGuid, err := match(h.authorizedAppGuids, req.TLS)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("client not authorized"))
		return
	}

	w.Write([]byte(fmt.Sprintf("success for instance %s of app %s", instanceGuid, appGuid)))
}

var ErrNoMatch = errors.New("failed to find an authorized app guid")

func match(appGuids []string, tlsInfo *tls.ConnectionState) (string, string, error) {
	if appGuids == nil {
		return "", "", nil
	}

	for _, appGuid := range appGuids {
		for _, cert := range tlsInfo.PeerCertificates {
			for _, ou := range cert.Subject.OrganizationalUnit {
				if strings.HasPrefix(ou, "app:") && ou[4:] == appGuid {
					return appGuid, cert.Subject.CommonName, nil
				}
			}
		}
	}

	return "", "", ErrNoMatch
}
