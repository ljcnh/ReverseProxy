// Copyright 2022 <mzh.scnu@qq.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package balancer

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

// TestBounded_Add .
func TestBounded_Add(t *testing.T) {
	expect, err := Build(ConsistentHashWithLoadBalancer, []string{"192.168.1.1:1015",
		"192.168.1.1:1016", "192.168.1.1:1017", "192.168.1.1:1018"})
	assert.Equal(t, err, nil)
	bounded := NewConsistentHashWithLoad(nil)
	bounded.Add("192.168.1.1:1015")
	bounded.Add("192.168.1.1:1016")
	bounded.Add("192.168.1.1:1017")
	bounded.Add("192.168.1.1:1018")
	tmp1 := expect.(*ConsistentHashWithLoad)
	tmp2 := bounded.(*ConsistentHashWithLoad)
	assert.Equal(t, true, SliceEql(tmp1.getHosts(), tmp2.getHosts()))
}

// TestBounded_Remove .
func TestBounded_Remove(t *testing.T) {
	expect, err := Build(ConsistentHashWithLoadBalancer, []string{"192.168.1.1:1015",
		"192.168.1.1:1016"})
	assert.Equal(t, err, nil)
	bounded := NewConsistentHashWithLoad([]string{"192.168.1.1:1015",
		"192.168.1.1:1016", "192.168.1.1:1017"})
	bounded.Remove("192.168.1.1:1017")
	tmp1 := expect.(*ConsistentHashWithLoad)
	tmp2 := bounded.(*ConsistentHashWithLoad)
	assert.Equal(t, true, SliceEql(tmp1.getHosts(), tmp2.getHosts()))
}

func TestBounded_Balance(t *testing.T) {
	expect, _ := Build(ConsistentHashWithLoadBalancer, []string{"192.168.1.1:1015",
		"192.168.1.1:1016", "192.168.1.1:1017", "192.168.1.1:1018"})
	expect.Inc("192.168.1.1:1015")
	expect.Inc("192.168.1.1:1015")
	expect.Inc("NIL")
	expect.Done("192.168.1.1:1015")
	expect.Done("NIL")
	host, _ := expect.Next("172.166.2.44")
	assert.Equal(t, "192.168.1.1:1017", host)
}

func SliceEql(s1, s2 []string) bool {
	mp1 := make(map[string]struct{})
	for _, s := range s1 {
		mp1[s] = struct{}{}
	}
	mp2 := make(map[string]struct{})
	for _, s := range s2 {
		mp2[s] = struct{}{}
	}
	return reflect.DeepEqual(mp1, mp2)
}
