package balancer

import (
	"sync"
)

func init() {
	AllBalanceAlgo[RoundRobinBalancer] = NewRoundRobin
}

type RoundRobin struct {
	mu    sync.RWMutex
	hosts []string
	index uint64
}

func NewRoundRobin(hosts []string) Balancer {
	return &RoundRobin{
		hosts: hosts,
		index: 0,
	}
}

func (round *RoundRobin) Add(host string) {
	round.mu.Lock()
	defer round.mu.Unlock()
	for _, h := range round.hosts {
		if h == host {
			return
		}
	}
	round.hosts = append(round.hosts, host)
}

func (round *RoundRobin) Remove(host string) {
	round.mu.Lock()
	defer round.mu.Unlock()
	for i, h := range round.hosts {
		if h == host {
			round.hosts = append(round.hosts[:i], round.hosts[i+1:]...)
		}
	}
}

func (round *RoundRobin) Next(_ string) (string, error) {
	round.mu.RLock()
	defer round.mu.RUnlock()
	if len(round.hosts) == 0 {
		return "", NoHost
	}
	pos := round.index
	round.index = (round.index + 1) % uint64(len(round.hosts))
	return round.hosts[pos], nil
}

func (round *RoundRobin) Inc(_ string) {
}

func (round *RoundRobin) Done(_ string) {
}

func (round *RoundRobin) Count() uint {
	round.mu.RLock()
	defer round.mu.RUnlock()
	return uint(len(round.hosts))
}

func (round *RoundRobin) Find(host string) bool {
	round.mu.RLock()
	defer round.mu.RUnlock()
	for _, h := range round.hosts {
		if h == host {
			return true
		}
	}
	return false
}
