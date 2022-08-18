package proxy

import (
	"fmt"
	"github.com/ljcnh/ReverseProxy/balancer"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

const ReverseProxy = "ReverseProxy"

var (
	// XRealIP XProxy XForwardedFor  CanonicalHeaderKey, 规范化string
	XRealIP       = http.CanonicalHeaderKey("X-Real-IP")
	XProxy        = http.CanonicalHeaderKey("X-Proxy")
	XForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
)

type HTTPProxy struct {
	hostMap  map[string]*httputil.ReverseProxy
	balancer balancer.Balancer
	mu       sync.RWMutex
	alive    map[string]ServerConnectStatus
}

func NewHTTPProxy(targetHosts []string, algorithm string) (*HTTPProxy, error) {
	hostMap := make(map[string]*httputil.ReverseProxy)
	alive := make(map[string]ServerConnectStatus)
	var hosts []string
	for _, targetHost := range targetHosts {
		url, err := url.Parse(targetHost)
		if err != nil {
			return nil, err
		}

		proxy := httputil.NewSingleHostReverseProxy(url)

		oriDirector := proxy.Director
		proxy.Director = func(request *http.Request) {
			oriDirector(request)
			request.Header.Set(XProxy, ReverseProxy)
			request.Header.Set(XRealIP, getIp(request))
		}

		host := getHost(url)
		hostMap[host] = proxy
		alive[host] = NORMAL
		hosts = append(hosts, host)
	}

	bl, err := balancer.Build(algorithm, hosts)
	if err != nil {
		return nil, err
	}
	return &HTTPProxy{
		hostMap:  hostMap,
		balancer: bl,
		alive:    alive,
	}, nil
}

func (proxy *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("proxy error: %s", err)
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(err.(error).Error()))
		}
	}()

	url, err := proxy.balancer.Next(getIp(r))
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(fmt.Sprintf("balance error: %s", err.Error())))
		return
	}
	proxy.balancer.Inc(url)
	defer proxy.balancer.Done(url)
	proxy.hostMap[url].ServeHTTP(w, r)
}

// HealthCheck

func (proxy *HTTPProxy) HealthCheck(interval uint) {
	for host := range proxy.hostMap {
		go proxy.healthCheck(host, interval)
	}
}

func (proxy *HTTPProxy) healthCheck(host string, interval uint) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for range ticker.C {
		status := isConnection(host)
		preStatus := proxy.readAlive(host)
		if preStatus != status {
			if status == NORMAL {
				proxy.balancer.Add(host)
			} else if preStatus == NORMAL {
				proxy.balancer.Remove(host)
			}
			proxy.setAlive(host, status)
		}
	}
}

func (proxy *HTTPProxy) setAlive(host string, status ServerConnectStatus) {
	proxy.mu.Lock()
	defer proxy.mu.Unlock()
	proxy.alive[host] = status
}

func (proxy *HTTPProxy) readAlive(url string) ServerConnectStatus {
	proxy.mu.RLocker()
	defer proxy.mu.RUnlock()
	return proxy.alive[url]
}
