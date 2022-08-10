package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

var (
	XRealIP       = http.CanonicalHeaderKey("X-Real-IP")
	XProxy        = http.CanonicalHeaderKey("X-Proxy")
	XForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
)

func GetIP(r *http.Request) string {
	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	fmt.Println("RemoteAddr", clientIP)
	if len(r.Header.Get(XForwardedFor)) != 0 {
		xff := r.Header.Get(XForwardedFor)
		s := strings.Index(xff, ", ")
		if s == -1 {
			s = len(r.Header.Get(XForwardedFor))
		}
		clientIP = xff[:s]
		fmt.Println("XForwardedFor", clientIP)
	} else if len(r.Header.Get(XRealIP)) != 0 {
		clientIP = r.Header.Get(XRealIP)
		fmt.Println("XRealIP", clientIP)
	}

	return clientIP
}

func ServeHTTP1(w http.ResponseWriter, r *http.Request) {
	log.Println("Server1")
	log.Println(r.URL)
	log.Println(r.RemoteAddr)
	GetIP(r)
	w.Write([]byte(string("Server1\n")))
}

func main() {
	http.HandleFunc("/", ServeHTTP1)
	http.ListenAndServe("127.0.0.1:1331", nil)
}
