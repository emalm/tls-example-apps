package main

import (
	"errors"
	"math/rand"
	"sync"

	"github.com/emalm/tls-example-apps/discovery"
)

var ErrNoLocations = errors.New("no backend locations stored")

type Backends struct {
	sync.RWMutex
	backendMap  map[string]discovery.Location
	backendList []discovery.Location
}

func NewBackends() *Backends {
	b := &Backends{
		backendMap: map[string]discovery.Location{},
	}
	b.unsafeRegenerateList()

	return b
}

func (b *Backends) Pick() (discovery.Location, error) {
	b.RLock()
	defer b.RUnlock()

	if len(b.backendList) == 0 {
		return discovery.Location{}, ErrNoLocations
	}

	return b.backendList[rand.Intn(len(b.backendList))], nil

}

func (b *Backends) Add(locs []discovery.Location) {
	b.Lock()
	defer b.Unlock()

	for _, loc := range locs {
		b.backendMap[loc.IPAddress] = loc
	}

	b.unsafeRegenerateList()
}

func (b *Backends) Remove(l discovery.Location) {
	b.Lock()
	defer b.Unlock()

	delete(b.backendMap, l.IPAddress)
	b.unsafeRegenerateList()
}

func (b *Backends) unsafeRegenerateList() {
	list := []discovery.Location{}

	for _, l := range b.backendMap {
		list = append(list, l)
	}

	b.backendList = list
}
