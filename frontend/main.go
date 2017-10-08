package main

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func main() {
	port := os.Getenv("PORT")

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

	handler := NewHandler(client)

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
	client *HTTPClient
}

func NewHandler(client *HTTPClient) *handler {
	return &handler{
		client: client,
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

	client := h.client.GetClient()

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
