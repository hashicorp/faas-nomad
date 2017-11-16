package handlers

import (
	"bytes"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// ProxyClient defines the interface for a client which calls faas functions
type ProxyClient interface {
	GetFunctionName(*http.Request) string
	CallAndReturnResponse(address string, body []byte, headers http.Header) ([]byte, http.Header, error)
}

// HTTPProxyClient allows the calling of functions
type HTTPProxyClient struct {
	proxyClient *http.Client
}

// MakeProxyClient creates a new HTTPProxyClient
func MakeProxyClient() *HTTPProxyClient {
	proxyClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 0,
			}).DialContext,
			MaxIdleConns:          200,
			DisableKeepAlives:     true,
			IdleConnTimeout:       120 * time.Millisecond,
			ExpectContinueTimeout: 1500 * time.Millisecond,
		},
	}

	return &HTTPProxyClient{
		proxyClient: proxyClient,
	}
}

// GetFunctionName returns the name of the function from the request vars
func (pc *HTTPProxyClient) GetFunctionName(r *http.Request) string {
	vars := mux.Vars(r)
	return vars["name"]
}

// CallAndReturnResponse calls the function and resturns the response
func (pc *HTTPProxyClient) CallAndReturnResponse(address string, body []byte, headers http.Header) (
	[]byte, http.Header, error) {
	stamp := strconv.FormatInt(time.Now().Unix(), 10)

	defer func(when time.Time) {
		seconds := time.Since(when).Seconds()
		log.Printf("[%s] took %f seconds\n", stamp, seconds)
	}(time.Now())

	log.Println("Trying to call:", address)
	request, _ := http.NewRequest("POST", address, bytes.NewReader(body))

	copyHeaders(&request.Header, &headers)

	defer request.Body.Close()

	response, err := pc.proxyClient.Do(request)
	if err != nil {
		log.Println(err.Error())
		return nil, nil, err
	}

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading body: %v\n", err)

		return nil, nil, err
	}
	response.Body.Close()

	log.Println("Finished")

	return respBody, response.Header, nil
}

func copyHeaders(destination *http.Header, source *http.Header) {
	for k, vv := range *source {
		vvClone := make([]string, len(vv))
		copy(vvClone, vv)
		(*destination)[k] = vvClone
	}
}

func randomInt(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
