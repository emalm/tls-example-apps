package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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
