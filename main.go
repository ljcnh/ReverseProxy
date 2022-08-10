package main

import (
	"github.com/gorilla/mux"
	"github.com/ljcnh/ReverseProxy/middleware"
	"github.com/ljcnh/ReverseProxy/proxy"
	"log"
	"net/http"
	"strconv"
)

// https://learnku.com/articles/37867
// https://golangr.com/golang-http-server/

func main() {
	config, err := ReadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Read the configuration file failed: %v \n", err)
		return
	}
	err = config.simpleValidation()
	if err != nil {
		log.Fatalf("Varify the configuration file failed: %v \n", err)
		return
	}
	router := mux.NewRouter()
	for _, l := range config.Location {
		httpProxy, err := proxy.NewHTTPProxy(l.ProxyPass, l.BalanceMode)
		if err != nil {
			log.Fatalf("create proxy failed: %v \n", err)
		}
		if config.HealthCheck {
			httpProxy.HealthCheck(config.HealthCheckInterval)
		}
		router.Handle(l.Pattern, httpProxy)
	}
	if config.MaxAllowed > 0 {
		router.Use(middleware.MaxAllowMiddleware(config.MaxAllowed))
	}
	ser := http.Server{
		Addr:    ":" + strconv.Itoa(config.Port),
		Handler: router,
	}
	if config.Schema == "http" {
		err = ser.ListenAndServe()
	} else if config.Schema == "https" {
		err = ser.ListenAndServeTLS(config.SSLCertificate, config.SSLCertificateKey)
	}
	if err != nil {
		log.Fatalf("listen and server failed: %v", err)
	}
}
