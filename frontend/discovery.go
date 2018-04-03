package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/emalm/tls-example-apps/discovery"
)

type LocationFinder interface {
	FindLocations() ([]discovery.Location, error)
}

func DiscoverBackends(b *Backends, finder LocationFinder, interval time.Duration) {
	timer := time.NewTimer(0 * time.Second)

	fmt.Printf("polling for backends every %s\n", interval)

	for {
		select {
		case <-timer.C:
			l, err := finder.FindLocations()
			if err == nil {
				b.Add(l)
			} else {
				fmt.Fprintf(os.Stderr, "error discovering backend: %s\n", err.Error())
			}

			timer.Reset(interval)
		}
	}
}

type RequestFinder struct {
	URL string
}

func (finder RequestFinder) FindLocations() ([]discovery.Location, error) {
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

	resp, err := client.Get(finder.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	location := discovery.Location{}
	err = json.Unmarshal(body, &location)

	return []discovery.Location{location}, err
}

type DNSFinder struct {
	Domain string
	Port   string
}

func (finder DNSFinder) FindLocations() ([]discovery.Location, error) {
	resolver := net.Resolver{}

	ips, err := resolver.LookupIPAddr(context.Background(), finder.Domain)
	if err != nil {
		return nil, err
	}

	locs := make([]discovery.Location, 0)

	for _, ip := range ips {
		if ip.IP.To4() != nil {
			locs = append(locs, discovery.Location{
				IPAddress: ip.IP.String(),
				TLSPort:   finder.Port,
			})
		}
	}

	return locs, nil
}
