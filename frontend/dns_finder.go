package main

import (
	"context"
	"net"

	"github.com/emalm/tls-example-apps/discovery"
)

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
