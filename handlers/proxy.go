package handlers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/nicholasjackson/faas-nomad/nomad"
)

// MakeProxy creates a proxy for HTTP web requests which can be routed to a function.
func MakeProxy(client nomad.Job) http.HandlerFunc {
	proxyClient := http.Client{
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

	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		if r.Method == "POST" {

			vars := mux.Vars(r)
			service := vars["name"]

			stamp := strconv.FormatInt(time.Now().Unix(), 10)

			defer func(when time.Time) {
				seconds := time.Since(when).Seconds()
				log.Printf("[%s] took %f seconds\n", stamp, seconds)
			}(time.Now())

			requestBody, _ := ioutil.ReadAll(r.Body)
			defer r.Body.Close()

			url := resolveFunction(client, service)
			log.Println("trying to call:", url)

			request, _ := http.NewRequest("POST", url, bytes.NewReader(requestBody))

			copyHeaders(&request.Header, &r.Header)

			defer request.Body.Close()

			response, err := proxyClient.Do(request)
			if err != nil {
				log.Println(err.Error())
				writeHead(service, http.StatusInternalServerError, w)
				buf := bytes.NewBufferString("Can't reach service: " + service)
				w.Write(buf.Bytes())
				return
			}

			clientHeader := w.Header()
			copyHeaders(&clientHeader, &response.Header)

			// TODO: copyHeaders removes the need for this line - test removal.
			// Match header for strict services
			w.Header().Set("Content-Type", r.Header.Get("Content-Type"))

			responseBody, _ := ioutil.ReadAll(response.Body)

			writeHead(service, http.StatusOK, w)
			w.Write(responseBody)

		}
	}
}

func resolveFunction(client nomad.Job, function string) string {
	job, _, err := client.Info(function, nil)
	if err != nil {
		log.Println("Error locating job")
	}

	return fmt.Sprintf(
		"http://docker.for.mac.localhost:%v",
		job.TaskGroups[0].Tasks[0].Config["port_map"].([]interface{})[0].(map[string]interface{})["http"].(float64),
	)
}

func writeHead(service string, code int, w http.ResponseWriter) {
	w.WriteHeader(code)
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
