package balancer

import (
	"github.com/ljcnh/ReverseProxy/consistent"
)

// 带有权重的封装

func init() {
	AllBalanceAlgo[ConsistentHashWithLoadBalancer] = NewConsistentHashWithLoad
}

type ConsistentHashWithLoad struct {
	consistentHash *consistent.Consistent
}

func NewConsistentHashWithLoad(hosts []string) Balancer {
	c := &ConsistentHashWithLoad{consistentHash: consistent.DefaultConsistent()}
	for _, h := range hosts {
		c.Add(h)
	}
	return c
}

func (c *ConsistentHashWithLoad) Add(host string) {
	c.consistentHash.Add(host)
}

func (c *ConsistentHashWithLoad) Remove(host string) {
	c.consistentHash.Remove(host)
}

func (c *ConsistentHashWithLoad) Next(key string) (string, error) {
	if c.Count() <= 0 {
		return "", NoHost
	}
	return c.consistentHash.GetLeast(key)
}

func (c *ConsistentHashWithLoad) Inc(host string) {
	c.consistentHash.Inc(host)
}

func (c *ConsistentHashWithLoad) Done(host string) {
	c.consistentHash.Done(host)
}

func (c *ConsistentHashWithLoad) Count() uint {
	return c.consistentHash.Count()
}

func (c *ConsistentHashWithLoad) Find(host string) bool {
	return c.consistentHash.Find(host)
}

func (c *ConsistentHashWithLoad) getHosts() []string {
	return c.consistentHash.GetHosts()
}
