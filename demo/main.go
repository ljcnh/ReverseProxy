package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

const PORT = "1330"
const A_CONDITION_URL = "http://localhost:1331"
const B_CONDITION_URL = "http://localhost:1332"
const DEFAULT_CONDITION_URL = "http://localhost:1333"

type requestPayloadStruct struct {
	ProxyCondition string `json:"proxy_condition"`
}

func getListenAddress() string {
	return ":" + PORT
}

func logSetup() {
	log.Printf("Server will run on: %s\n", getListenAddress())
	log.Printf("Redirecting to A url: %s\n", A_CONDITION_URL)
	log.Printf("Redirecting to B url: %s\n", B_CONDITION_URL)
	log.Printf("Redirecting to Default url: %s\n", DEFAULT_CONDITION_URL)
}

func requestBodyDecoder(request *http.Request) *json.Decoder {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		panic(err)
	}
	request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(body)))
}

func parseRequestBody(request *http.Request) requestPayloadStruct {
	decoder := requestBodyDecoder(request)
	var requestPayload requestPayloadStruct
	err := decoder.Decode(&requestPayload)
	if err != nil {
		panic(err)
	}
	return requestPayload
}

func getProxyUrl(proxyConditionRaw string) string {
	proxyCondition := strings.ToUpper(proxyConditionRaw)
	if proxyCondition == "A" {
		return A_CONDITION_URL
	}
	if proxyCondition == "B" {
		return B_CONDITION_URL
	}
	return DEFAULT_CONDITION_URL
}

func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	url, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(res, req)
}

func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	requestPayload := parseRequestBody(req)
	url := getProxyUrl(requestPayload.ProxyCondition)
	log.Printf("proxy_condition: %s, proxy_url: %s\n", requestPayload.ProxyCondition, url)
	serveReverseProxy(url, res, req)
	log.Println(req.URL)
	log.Println(req.RemoteAddr)
}

func main() {
	logSetup()
	http.HandleFunc("/", handleRequestAndRedirect)
	if err := http.ListenAndServe(getListenAddress(), nil); err != nil {
		panic(err)
	}
}
