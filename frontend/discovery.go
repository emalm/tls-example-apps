package main

import (
	"fmt"
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
