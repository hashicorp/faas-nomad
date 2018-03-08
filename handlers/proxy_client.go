package handlers

import (
	"bytes"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	hclog "github.com/hashicorp/go-hclog"
)

// ProxyClient defines the interface for a client which calls faas functions
type ProxyClient interface {
	GetFunctionName(*http.Request) string
	CallAndReturnResponse(address string, body []byte, headers http.Header) ([]byte, http.Header, error)
}

// HTTPProxyClient allows the calling of functions
type HTTPProxyClient struct {
	proxyClient *http.Client
	logger      hclog.Logger
}

// MakeProxyClient creates a new HTTPProxyClient
func MakeProxyClient(timeout time.Duration, l hclog.Logger) *HTTPProxyClient {
	proxyClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
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
		logger:      l,
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

	defer func(when time.Time) {
		seconds := time.Since(when).Seconds()
		pc.logger.Info("Execution time", "address", address, "duration(s)", seconds)
	}(time.Now())

	pc.logger.Info("Trying to call:", "address", address)
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
		pc.logger.Error("Error reading body", "error", err)

		return nil, nil, err
	}
	response.Body.Close()

	pc.logger.Info("Finished")

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
