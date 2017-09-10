package handlers

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type ProxyClient interface {
	GetFunctionName(*http.Request) string
	CallAndReturnResponse(string, http.ResponseWriter, *http.Request)
}

type HTTPProxyClient struct {
	proxyClient *http.Client
}

func MakeProxyClient() *HTTPProxyClient {
	proxyClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 0,
			}).DialContext,
			MaxIdleConns:          1,
			DisableKeepAlives:     true,
			IdleConnTimeout:       120 * time.Millisecond,
			ExpectContinueTimeout: 1500 * time.Millisecond,
		},
	}

	return &HTTPProxyClient{
		proxyClient: proxyClient,
	}
}

func (pc *HTTPProxyClient) GetFunctionName(r *http.Request) string {
	vars := mux.Vars(r)
	return vars["name"]
}

func (pc *HTTPProxyClient) CallAndReturnResponse(address string, w http.ResponseWriter, r *http.Request) {
	stamp := strconv.FormatInt(time.Now().Unix(), 10)

	defer func(when time.Time) {
		seconds := time.Since(when).Seconds()
		log.Printf("[%s] took %f seconds\n", stamp, seconds)
	}(time.Now())

	requestBody, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	log.Println("Trying to call:", address)
	request, _ := http.NewRequest("POST", address, bytes.NewReader(requestBody))

	copyHeaders(&request.Header, &r.Header)

	defer request.Body.Close()

	response, err := pc.proxyClient.Do(request)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	clientHeader := w.Header()
	copyHeaders(&clientHeader, &response.Header)

	// TODO: copyHeaders removes the need for this line - test removal.
	// Match header for strict services
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	io.Copy(w, response.Body)
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
