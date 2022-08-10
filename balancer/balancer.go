package balancer

import "errors"

var (
	AlgoNotSupported = errors.New("algorithm not supported")
)

type Balancer interface {
	Add(string)
	Remove(string)
	Balance(string) (string, error)
	Inc(string)
	Done(string)
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
