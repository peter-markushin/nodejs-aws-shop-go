package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
)

var proxyCache cache.Cache

type CachingTransport struct{}

func shouldCacheURL(method string, url *url.URL) bool {
	if strings.ToLower(method) != "get" {
		return false
	}

	if !strings.Contains(url.Path, "/products") {
		return false
	}

	return true
}

func (CachingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	hash := sha1.Sum([]byte(r.URL.String()))
	cacheKey := string(hash[:])

	if x, found := proxyCache.Get(cacheKey); found {
		respDump := x.([]byte)

		return http.ReadResponse(bufio.NewReader(bytes.NewReader(respDump)), r)
	}

	resp, err := http.DefaultTransport.RoundTrip(r)

	if err == nil && shouldCacheURL(r.Method, r.URL) {
		dump, _ := httputil.DumpResponse(resp, true)
		proxyCache.Add(cacheKey, dump, cache.DefaultExpiration)
	}

	return resp, err
}

var proxyUrls map[string]string

func getProxyURL(req *http.Request) (*string, error) {
	// path := req.URL.Path
	parts := strings.Split(req.URL.Path, "/")

	serviceName := parts[1]

	if len(serviceName) == 0 {
		return nil, errors.New("incorrect path")
	}

	proxy, pOk := proxyUrls[serviceName]

	if !pOk {
		return nil, errors.New("unknown service")
	}

	return &proxy, nil
}

func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	url, _ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Transport = CachingTransport{}
	req.Host = req.URL.Host

	// Note that ServeHttp is non-blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}

// Given a request send it to the appropriate url
func handleRequestAndForward(res http.ResponseWriter, req *http.Request) {
	proxyUrl, err := getProxyURL(req)

	if err != nil {
		res.WriteHeader(502)

		return
	}

	serveReverseProxy(*proxyUrl, res, req)
}

func main() {
	proxyUrls = map[string]string{
		"profile":  os.Getenv("CART_BASE_URL"),
		"products": os.Getenv("PRODUCTS_BASE_URL"),
	}
	proxyCache = *cache.New(120*time.Second, 5*time.Second)

	http.HandleFunc("/", handleRequestAndForward)
	log.Fatalln(http.ListenAndServe(`:80`, nil))
}
