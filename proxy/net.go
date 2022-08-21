package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const ConnectionTimeout = 3 * time.Second

type ServerConnectStatus string

// tcp成功 ping一定成功  我感觉是这样...
const (
	NORMAL ServerConnectStatus = "Normal"
	// TCPFAILED 测试服务器端口
	TCPFAILED = "TcpFailed"
	// PINGFAILED 测试网络情况(ping)
	PINGFAILED = "PingFailed"
)

// getIp
// https://www.cnblogs.com/mypath/articles/5239687.html
// 使用代理  通过RemoteAddr得到的是代理服务器的Ip 需要通过XForwardedFor和XRealIP
// 得到客户端的真实Ip
func getIp(request *http.Request) string {
	clientIp, _, _ := net.SplitHostPort(request.RemoteAddr)
	if len(request.Header.Get(XForwardedFor)) > 0 {
		XFF := request.Header.Get(XForwardedFor)
		fmt.Println(XFF)
		//	取得第一个ip
		pos := strings.Index(XFF, ", ")
		if pos == -1 {
			pos = len(XFF)
		}
		clientIp = XFF[:pos]
	} else if len(request.Header.Get(XRealIP)) != 0 {
		clientIp = request.Header.Get(XRealIP)
	}
	return clientIp
}

func getHost(url *url.URL) string {
	if url.Scheme == "http" {
		return fmt.Sprintf("%s:%s", url.Host, "80")
	} else if url.Scheme == "https" {
		return fmt.Sprintf("%s:%s", url.Host, "443")
	} else {
		return url.Host
	}
}

// 检查服务器的网络连接是否正常
// isConnection  method: tcp或者ping
func isConnection(host string) (status ServerConnectStatus) {
	if tcpConnection(host) {
		status = NORMAL
	} else if netStatus(host) {
		status = TCPFAILED
	} else {
		status = PINGFAILED
	}
	return
}

func tcpConnection(host string) bool {
	tcpAddr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		return false
	}
	addr := fmt.Sprintf("%s:%d", tcpAddr.IP, tcpAddr.Port)
	conn, err := net.DialTimeout("tcp", addr, ConnectionTimeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// NetStatus only supoort
func netStatus(host string) bool {
	sysType := runtime.GOOS
	var command interface{}
	if sysType == "windows" {
		command = exec.Command("ping", host, "-n", "1")
	} else if sysType == "linux" {
		command = exec.Command("ping", host, "-c", "1", "-W", "5")
	}
	cmd := command.(*exec.Cmd)
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}
