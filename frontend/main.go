package main

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const defaultPort = "8080"
const defaultBackendPort = "9999"
const defaultBackendDiscoveryURL = "http://127.0.0.1:8080"
const discoveryPollingInterval = 5 * time.Second

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
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

	client := &HTTPClient{}

	go client.ResetClientPeriodically(certFilePath, keyFilePath, certPool, intervalDuration)

	backends := NewBackends()

	discoveryURL := os.Getenv("BACKEND_DISCOVERY_URL")
	if discoveryURL == "" {
		discoveryURL = defaultBackendDiscoveryURL
	}

	go DiscoverBackends(backends, discoveryURL, discoveryPollingInterval)

	handler := NewHandler(client, backends)

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
	client   *HTTPClient
	backends *Backends
}

func NewHandler(client *HTTPClient, backends *Backends) *handler {
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

	client := h.client.GetClient()

	resp, err := client.Get("https://" + ip + ":" + port)
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
		w.Write([]byte("backend " + location.InstanceGuid + " failed to authorize frontend: "))
		w.Write(body)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success from backend " + location.InstanceGuid + ": "))
	w.Write(body)
}
