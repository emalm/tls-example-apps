package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/emalm/tls-example-apps/discovery"
)

func DiscoverBackends(b *Backends, discoveryURL string, interval time.Duration) {
	timer := time.NewTimer(0 * time.Second)

	fmt.Printf("polling for backends every %s\n", interval)

	for {
		select {
		case <-timer.C:
			l, err := findLocation(discoveryURL)
			if err == nil {
				b.Add(l)
			} else {
				fmt.Fprintf(os.Stderr, "error discovering backend: %s\n", err.Error())
			}

			timer.Reset(interval)
		}
	}
}

func findLocation(discoveryURL string) (discovery.Location, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	location := discovery.Location{}

	resp, err := client.Get(discoveryURL)
	if err != nil {
		return location, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return location, err
	}

	err = json.Unmarshal(body, &location)

	return location, err
}
