package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func main() {
	port := os.Getenv("PORT")
	certFilePath := os.Getenv("CF_INSTANCE_CERT")
	keyFilePath := os.Getenv("CF_INSTANCE_KEY")
	cert, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)
	if err != nil {
		fmt.Printf("error reading cert from '%s' and '%s': %s\n", certFilePath, keyFilePath, err.Error())
		os.Exit(2)
	}

	caCertFilePath := os.Getenv("CA_CERT_FILE")
	var certPool *x509.CertPool
	if caCertFilePath != "" {
		certPool = x509.NewCertPool()
		certData, err := ioutil.ReadFile(caCertFilePath)
		if err != nil {
			fmt.Printf("error reading CA cert from '%s': %s\n", caCertFilePath, err.Error())
			os.Exit(2)

		}
		certPool.AppendCertsFromPEM(certData)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
		ClientAuth:   tls.RequireAnyClientCert,
	}

	authorizedAppGuidList := os.Getenv("AUTHORIZED_APP_GUIDS")
	handler, err := NewHandler(authorizedAppGuidList)
	if err != nil {
		fmt.Printf("error initializing handler from app guid list '%s'\n", authorizedAppGuidList)
		os.Exit(2)
	}

	server := http.Server{
		Addr:      "0.0.0.0:" + port,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	err = server.ListenAndServeTLS("", "")
	if err != nil {
		fmt.Printf("error serving: %s\n", err.Error())
	}
}

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
