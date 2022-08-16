package balancer

import (
	"math/rand"
	"sync"
	"time"
)

func init() {
	AllBalanceAlgo[RandomBalancer] = NewRandom
}

type Random struct {
	mu    sync.RWMutex
	hosts []string
	r     *rand.Rand
}

func NewRandom(hosts []string) Balancer {
	return &Random{
		hosts: hosts,
		r:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (random *Random) Add(host string) {
	random.mu.Lock()
	defer random.mu.Unlock()
	for _, h := range random.hosts {
		if h == host {
			return
		}
	}
	random.hosts = append(random.hosts, host)
}

func (random *Random) Remove(host string) {
	random.mu.Lock()
	defer random.mu.Unlock()
	for i, h := range random.hosts {
		if h == host {
			random.hosts = append(random.hosts[:i], random.hosts[i+1:]...)
			return
		}
	}
}

func (random *Random) Next(_ string) (string, error) {
	random.mu.RLock()
	defer random.mu.RUnlock()
	if len(random.hosts) == 0 {
		return "", NoHost
	}
	return random.hosts[random.r.Intn(len(random.hosts))], nil
}

func (random *Random) Inc(_ string) {
}

func (random *Random) Done(_ string) {
}

func (random *Random) Count() int {
	random.mu.RLock()
	defer random.mu.RUnlock()
	return len(random.hosts)
}

func (random *Random) Find(host string) bool {
	random.mu.RLock()
	defer random.mu.RUnlock()
	for _, h := range random.hosts {
		if h == host {
			return true
		}
	}
	return false
}
