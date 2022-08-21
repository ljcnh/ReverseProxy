package balancer

import (
	"github.com/ljcnh/ReverseProxy/consistent"
)

// 没有权重的封装

func init() {
	AllBalanceAlgo[ConsistentHashBalancer] = NewConsistentHash
}

type ConsistentHash struct {
	consistentHash *consistent.Consistent
}

func NewConsistentHash(hosts []string) Balancer {
	c := &ConsistentHash{consistentHash: consistent.DefaultConsistent()}
	for _, h := range hosts {
		c.Add(h)
	}
	return c
}

func (c *ConsistentHash) Add(host string) {
	c.consistentHash.Add(host)
}

func (c *ConsistentHash) Remove(host string) {
	c.consistentHash.Remove(host)
}

func (c *ConsistentHash) Next(key string) (string, error) {
	if c.Count() <= 0 {
		return "", NoHost
	}
	return c.consistentHash.Get(key)
}

func (c *ConsistentHash) Inc(_ string) {
}

func (c *ConsistentHash) Done(_ string) {
}

func (c *ConsistentHash) Count() uint {
	return c.consistentHash.Count()
}

func (c *ConsistentHash) Find(host string) bool {
	return c.consistentHash.Find(host)
}
