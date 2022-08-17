package balancer

import (
	"hash/crc32"
	"sync"
)

func init() {
	AllBalanceAlgo[IpHashBalancer] = NewIpHash
}

type Hasher func(data string) uint32

type IpHash struct {
	hasher Hasher // default  crc32
	hosts  []string
	mu     sync.RWMutex
}

func NewIpHash(hosts []string) Balancer {
	return &IpHash{
		hosts: hosts,
		hasher: func(data string) uint32 {
			return crc32.ChecksumIEEE([]byte(data))
		},
	}
}

func (ipHash *IpHash) Add(host string) {
	ipHash.mu.Lock()
	defer ipHash.mu.Unlock()
	for _, h := range ipHash.hosts {
		if h == host {
			return
		}
	}
	ipHash.hosts = append(ipHash.hosts, host)
}

func (ipHash *IpHash) Remove(host string) {
	ipHash.mu.Lock()
	defer ipHash.mu.Unlock()
	for i, h := range ipHash.hosts {
		if h == host {
			ipHash.hosts = append(ipHash.hosts[:i], ipHash.hosts[i+1:]...)
			return
		}
	}
}

func (ipHash *IpHash) Next(key string) (string, error) {
	ipHash.mu.RLock()
	defer ipHash.mu.RUnlock()
	if len(ipHash.hosts) == 0 {
		return "", NoHost
	}
	value := ipHash.hasher(key) % uint32(len(ipHash.hosts))
	return ipHash.hosts[value], nil
}

func (ipHash *IpHash) Inc(_ string) {
}

func (ipHash *IpHash) Done(_ string) {
}

func (ipHash *IpHash) Count() int {
	ipHash.mu.RLock()
	defer ipHash.mu.RUnlock()
	return len(ipHash.hosts)
}

func (ipHash *IpHash) Find(host string) bool {
	ipHash.mu.RLock()
	defer ipHash.mu.RUnlock()
	for _, h := range ipHash.hosts {
		if h == host {
			return true
		}
	}
	return false
}

func (ipHash *IpHash) getHosts() []string {
	ipHash.mu.RLock()
	defer ipHash.mu.RUnlock()
	return ipHash.hosts
}
