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

	"github.com/emalm/tls-example-apps/certs"
)

const defaultPort = "8080"
const defaultBackendPort = "9999"
const defaultBackendDomain = "127.0.0.1"
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

	certificate := &certs.Certificate{}
	go certificate.ResetCertificatePeriodically(certFilePath, keyFilePath, intervalDuration)

	tlsConfig := &tls.Config{
		GetClientCertificate: certificate.GetClientCertificate,
		RootCAs:              certPool,
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
			TLSClientConfig:       tlsConfig,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	backends := NewBackends()
	finder := constructLocationFinder()

	go DiscoverBackends(backends, finder, discoveryPollingInterval)

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

func constructLocationFinder() LocationFinder {
	if os.Getenv("USE_PLATFORM_SERVICE_DISCOVERY") != "true" {
		discoveryURL := os.Getenv("BACKEND_DISCOVERY_URL")
		if discoveryURL == "" {
			discoveryURL = defaultBackendDiscoveryURL
		}

		return RequestFinder{URL: discoveryURL}
	} else {
		domain := os.Getenv("BACKEND_DOMAIN")
		if domain == "" {
			domain = defaultBackendDomain
		}

		port := os.Getenv("BACKEND_PORT")
		if port == "" {
			port = defaultBackendPort
		}

		return DNSFinder{
			Domain: domain,
			Port:   port,
		}
	}
}
