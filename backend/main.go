package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const defaultPort = "8080"
const defaultTLSPort = "9999"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	tlsPort := os.Getenv("TLS_PORT")
	if tlsPort == "" {
		tlsPort = defaultTLSPort
	}

	certFilePath := os.Getenv("CF_INSTANCE_CERT")
	keyFilePath := os.Getenv("CF_INSTANCE_KEY")
	interval := os.Getenv("CERT_RELOAD_INTERVAL")
	if interval == "" {
		interval = "5m"
	}
	intervalDuration, err := time.ParseDuration(interval)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse duration '%s'\n", interval)
		os.Exit(2)
	}

	certificate := &Certificate{}
	go certificate.ResetCertificatePeriodically(certFilePath, keyFilePath, intervalDuration)

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

	instanceGuid := os.Getenv("CF_INSTANCE_GUID")
	ipAddress := os.Getenv("CF_INSTANCE_INTERNAL_IP")
	discovery := NewDiscoveryHandler(instanceGuid, ipAddress, tlsPort)

	server := http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: discovery,
	}

	go server.ListenAndServe()

	tlsConfig := &tls.Config{
		GetCertificate: certificate.GetCertificate,
		RootCAs:        certPool,
		ClientAuth:     tls.RequireAnyClientCert,
	}

	authorizedAppGuidList := os.Getenv("AUTHORIZED_APP_GUIDS")
	handler, err := NewHandler(authorizedAppGuidList)
	if err != nil {
		fmt.Printf("error initializing handler from app guid list '%s'\n", authorizedAppGuidList)
		os.Exit(2)
	}

	tlsServer := http.Server{
		Addr:      "0.0.0.0:" + tlsPort,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	err = tlsServer.ListenAndServeTLS("", "")
	if err != nil {
		fmt.Printf("error serving: %s\n", err.Error())
	}
}
