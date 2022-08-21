package consistent

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/minio/blake2b-simd"
	"math"
	"sort"
	"sync"
	"sync/atomic"
)

const replicationFactor = 10

var NoHosts = errors.New("no hosts added")

type Hash func(data string) uint64

type Host struct {
	Name string
	Load int64
}

type Consistent struct {
	hosts     map[uint64]string
	sorted    []uint64
	hostMap   map[string]*Host
	replicas  int
	totalLoad int64
	hashFunc  Hash

	mu sync.RWMutex
}

func New(replicas int, hash Hash) *Consistent {
	return &Consistent{
		hosts:    map[uint64]string{},
		sorted:   []uint64{},
		hostMap:  make(map[string]*Host),
		replicas: replicas,
		hashFunc: hash,
	}
}

func DefaultConsistent() *Consistent {
	return New(replicationFactor, hash)
}

func (c *Consistent) Add(host string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.hostMap[host]; ok {
		return
	}
	c.hostMap[host] = &Host{
		Name: host,
		Load: 0,
	}
	for i := 0; i < c.replicas; i++ {
		h := c.hashFunc(fmt.Sprintf("%s%d", host, i))
		c.hosts[h] = host
		c.sorted = append(c.sorted, h)
	}
	sort.Slice(c.sorted, func(i, j int) bool {
		if c.sorted[i] < c.sorted[j] {
			return true
		}
		return false
	})
}

func (c *Consistent) Get(key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.hosts) == 0 {
		return "", NoHosts
	}

	h := c.hashFunc(key)
	index := c.search(h)
	return c.hosts[c.sorted[index]], nil
}

func (c *Consistent) GetLeast(key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.hosts) == 0 {
		return "", NoHosts
	}
	h := c.hashFunc(key)
	index := c.search(h)

	i := index
	for {
		host := c.hosts[c.sorted[i]]
		if c.loadOK(host) {
			return host, nil
		}
		i++
		if i >= len(c.hosts) {
			i = 0
		}
	}
}

func (c *Consistent) UpdateLoad(host string, load int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.hostMap[host]; !ok {
		return
	}
	c.totalLoad -= c.hostMap[host].Load
	c.hostMap[host].Load = load
	c.totalLoad += load
}

func (c *Consistent) Inc(host string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.hostMap[host]; !ok {
		return
	}
	atomic.AddInt64(&c.hostMap[host].Load, 1)
	atomic.AddInt64(&c.totalLoad, 1)
	//c.hostMap[host].Load += 1
	//c.totalLoad += 1
}

func (c *Consistent) Done(host string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.hostMap[host]; !ok {
		return
	}
	atomic.AddInt64(&c.hostMap[host].Load, -1)
	atomic.AddInt64(&c.totalLoad, -1)
	//c.hostMap[host].Load -= 1
	//c.totalLoad -= 1
}

func (c *Consistent) Remove(host string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := 0; i < c.replicas; i++ {
		h := c.hashFunc(fmt.Sprintf("%s%d", host, i))
		delete(c.hosts, h)
		c.delSlice(h)
	}
	delete(c.hostMap, host)
	return true
}

func (c *Consistent) GetHosts() (hosts []string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for host, _ := range c.hostMap {
		hosts = append(hosts, host)
	}
	return hosts
}

func (c *Consistent) Count() uint {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return uint(len(c.hostMap))
}

func (c *Consistent) Find(host string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if _, ok := c.hostMap[host]; ok {
		return true
	}
	return false
}

func (c *Consistent) GetLoads() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	loads := map[string]int64{}
	for k, v := range c.hostMap {
		loads[k] = v.Load
	}
	return loads
}

// MaxLoad 返回单个host的最大load
// 不知道为啥  网上是这么计算的
// (total_load/number_of_hosts)*1.25
func (c *Consistent) MaxLoad() int64 {
	if c.totalLoad == 0 {
		c.totalLoad = 1
	}
	var avgLoad float64
	avgLoad = float64(c.totalLoad / int64(len(c.hostMap)))
	if avgLoad == 0 {
		avgLoad = 1
	}
	avgLoad = math.Ceil(avgLoad * 1.25)
	return int64(avgLoad)
}

func (c *Consistent) loadOK(host string) bool {
	if c.totalLoad < 0 {
		c.totalLoad = 0
	}
	var avgLoad float64
	avgLoad = float64((c.totalLoad + 1) / int64(len(c.hostMap)))
	if avgLoad == 0 {
		avgLoad = 1
	}
	avgLoad = math.Ceil(avgLoad * 1.25)
	hHost, ok := c.hostMap[host]
	if !ok {
		panic(fmt.Sprintf("given host(%s) not in loadsMap", hHost.Name))
	}
	if float64(hHost.Load)+1 <= avgLoad {
		return true
	}
	return false
}

func (c *Consistent) search(key uint64) int {
	idx := sort.Search(len(c.sorted), func(i int) bool {
		return c.sorted[i] >= key
	})
	if idx >= len(c.sorted) {
		return 0
	}
	return idx
}

func (c *Consistent) delSlice(val uint64) {
	idx := -1
	left := 0
	right := len(c.sorted) - 1
	for left <= right {
		mid := (right + left) / 2
		if c.sorted[mid] == val {
			idx = mid
			break
		} else if c.sorted[mid] > val {
			right = mid - 1
		} else {
			left = mid + 1
		}
	}
	if idx != -1 {
		c.sorted = append(c.sorted[:idx], c.sorted[idx+1:]...)
	}
}

func hash(host string) uint64 {
	out := blake2b.Sum512([]byte(host))
	return binary.LittleEndian.Uint64(out[:])
}
