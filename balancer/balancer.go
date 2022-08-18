package balancer

import "errors"

var (
	AlgoNotSupported = errors.New("algorithm not supported")
	NoHost           = errors.New("no host")
)

type Balancer interface {
	Add(string)
	Remove(string)
	Next(string) (string, error)
	Inc(string)
	Done(string)
	Count() uint
	Find(string) bool
}

type BalanceAlgo func([]string) Balancer

var AllBalanceAlgo = make(map[string]BalanceAlgo)

func Build(algorithm string, hosts []string) (Balancer, error) {
	balanceAlgo, ok := AllBalanceAlgo[algorithm]
	if !ok {
		return nil, AlgoNotSupported
	}
	return balanceAlgo(hosts), nil
}
