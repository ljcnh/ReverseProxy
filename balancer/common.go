package balancer

const (
	RoundRobinBalancer = "round-robin"
	RandomBalancer     = "random"
	IpHashBalancer     = "ip-hash"
	P2CBalancer        = "p2c"
	LeastLoadBalancer  = "least-load"
)

// hostLoad host-load 网络-负载(连接数)
type hostLoad struct {
	host string
	load uint64
}

func (s *hostLoad) Tag() interface{} {
	return s.host
}

func (s *hostLoad) Key() float64 {
	return float64(s.load)
}
