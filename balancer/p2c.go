package balancer

import (
	"hash/crc32"
	"math/rand"
	"sync"
	"time"
)

const Salt = "%#!"

func init() {
	AllBalanceAlgo[P2CBalancer] = NewP2C
}

type P2C struct {
	mu      sync.RWMutex
	r       *rand.Rand
	hosts   []*hostLoad
	hostMap map[string]*hostLoad
}

func NewP2C(hosts []string) Balancer {
	p2c := &P2C{
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
		hosts:   []*hostLoad{},
		hostMap: make(map[string]*hostLoad),
	}
	for _, h := range hosts {
		p2c.Add(h)
	}
	return p2c
}

func (p *P2C) Add(host string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.hostMap[host]; ok {
		return
	}
	h := &hostLoad{
		host: host,
		load: 0,
	}
	p.hosts = append(p.hosts, h)
	p.hostMap[host] = h
}

func (p *P2C) Remove(host string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.hostMap[host]; !ok {
		return
	}
	delete(p.hostMap, host)
	for i, h := range p.hosts {
		if h.host == host {
			p.hosts = append(p.hosts[:i], p.hosts[i+1:]...)
			return
		}
	}
}

func (p *P2C) Next(key string) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(p.hosts) == 0 {
		return "", NoHost
	}
	host := p.choose(key)
	return host, nil
}

func (p *P2C) choose(key string) string {
	var c1, c2 string
	if len(key) > 0 {
		saltKey := key + Salt
		c1 = p.hosts[p.crcHash(key)].host
		c2 = p.hosts[p.crcHash(saltKey)].host
	} else {
		c1 = p.hosts[p.r.Intn(len(p.hosts))].host
		c2 = p.hosts[p.r.Intn(len(p.hosts))].host
	}
	if p.hostMap[c1].load <= p.hostMap[c2].load {
		return c1
	}
	return c2
}

func (p *P2C) crcHash(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key)) % uint32(len(p.hosts))
}

// Inc load+1
func (p *P2C) Inc(host string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if h, ok := p.hostMap[host]; ok {
		h.load++
	}
}

// Done load-1
func (p *P2C) Done(host string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if h, ok := p.hostMap[host]; ok {
		if h.load > 0 {
			h.load--
		}
	}
}

func (p *P2C) Count() uint {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return uint(len(p.hosts))
}

func (p *P2C) Find(host string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if _, ok := p.hostMap[host]; ok {
		return true
	}
	return false
}
