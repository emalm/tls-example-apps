package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
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
	}

	handler := NewHandler(tlsConfig)

	server := http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: handler,
	}

	err = server.ListenAndServe()
	if err != nil {
		fmt.Printf("error serving: %s\n", err.Error())
	}
}

type handler struct {
	tlsConfig *tls.Config
}

func NewHandler(tlsConfig *tls.Config) *handler {
	return &handler{
		tlsConfig: tlsConfig,
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error parsing request: " + err.Error()))
		return
	}

	ip := req.FormValue("ip")
	if ip == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error: no 'ip' field in request"))
		return
	}

	port := req.FormValue("port")
	if port == "" {
		port = "8080"
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			TLSClientConfig:       h.tlsConfig,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	resp, err := client.Get("https://" + ip + ":" + port)
	if err != nil {
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
		w.Write([]byte("backend failed to authorize frontend: "))
		w.Write(body)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success: "))
	w.Write(body)
}
