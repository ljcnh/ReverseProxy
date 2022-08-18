package balancer

import (
	fibHeap "github.com/starwander/GoFibonacciHeap"
	"sync"
)

// https://github.com/starwander/GoFibonacciHeap

func init() {
	AllBalanceAlgo[LeastLoadBalancer] = NewLeastLoad
}

type LeastLoad struct {
	mu   sync.RWMutex
	heap *fibHeap.FibHeap
}

func NewLeastLoad(hosts []string) Balancer {
	ll := &LeastLoad{heap: fibHeap.NewFibHeap()}
	for _, host := range hosts {
		ll.Add(host)
	}
	return ll
}

func (ll *LeastLoad) Add(host string) {
	ll.mu.Lock()
	defer ll.mu.Unlock()
	val := ll.heap.GetValue(host)
	if val != nil {
		return
	}
	_ = ll.heap.InsertValue(&hostLoad{
		host: host,
		load: 0,
	})
}

func (ll *LeastLoad) Remove(host string) {
	ll.mu.Lock()
	defer ll.mu.Unlock()
	val := ll.heap.GetValue(host)
	if val == nil {
		return
	}
	_ = ll.heap.Delete(host)
}

func (ll *LeastLoad) Next(_ string) (string, error) {
	ll.mu.RLock()
	defer ll.mu.RUnlock()
	if ll.heap.Num() == 0 {
		return "", NoHost
	}
	return ll.heap.MinimumValue().Tag().(string), nil
}

func (ll *LeastLoad) Inc(host string) {
	ll.mu.Lock()
	defer ll.mu.Unlock()
	if ok := ll.heap.GetValue(host); ok == nil {
		return
	}
	val := ll.heap.GetValue(host)
	val.(*hostLoad).load += 1
	_ = ll.heap.IncreaseKeyValue(val)
}

func (ll *LeastLoad) Done(host string) {
	ll.mu.Lock()
	defer ll.mu.Unlock()
	val := ll.heap.GetValue(host)
	if ok := ll.heap.GetValue(host); ok == nil {
		return
	}
	if val.(*hostLoad).load > 0 {
		val.(*hostLoad).load -= 1
	}
	_ = ll.heap.IncreaseKeyValue(val)
}

func (ll *LeastLoad) Count() uint {
	ll.mu.RLock()
	defer ll.mu.RUnlock()
	return ll.heap.Num()
}

func (ll *LeastLoad) Find(host string) bool {
	ll.mu.RLock()
	defer ll.mu.RUnlock()
	if ok := ll.heap.GetValue(host); ok != nil {
		return true
	}
	return false
}
