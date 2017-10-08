package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

type HTTPClient struct {
	client *http.Client
	sync.Mutex
}

func (c *HTTPClient) GetClient() *http.Client {
	c.Lock()
	defer c.Unlock()
	return c.client
}

func (c *HTTPClient) setClient(client *http.Client) {
	c.Lock()
	defer c.Unlock()
	c.client = client
}

func (c *HTTPClient) ResetClientPeriodically(certFilePath, keyFilePath string, certPool *x509.CertPool, interval time.Duration) {
	timer := time.NewTimer(0 * time.Second)

	fmt.Printf("reloading certificates every %s\n", interval)

	for {
		select {
		case <-timer.C:
			certificate, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading cert from '%s' and '%s': %s\n", certFilePath, keyFilePath, err.Error())
				os.Exit(2)
			}

			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{certificate},
				RootCAs:      certPool,
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

			c.setClient(client)
			fmt.Printf("reloaded certificate at %v\n", time.Now())

			timer.Reset(interval)
		}
	}

}
