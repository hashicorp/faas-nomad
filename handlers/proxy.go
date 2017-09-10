package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/faas-nomad/consul"
)

// MakeProxy creates a proxy for HTTP web requests which can be routed to a function.
func MakeProxy(client ProxyClient, resolver consul.ServiceResolver) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		service := client.GetFunctionName(r)
		if service == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Expected POST")
			return
		}

		urls, _ := resolver.Resolve(service)

		// hack for docker for mac, need real implementation
		address := urls[0]
		if strings.Contains(address, "127.0.0.1") {
			address = strings.Replace(address, "127.0.0.1", "docker.for.mac.localhost", 1)
		}

		client.CallAndReturnResponse(address, w, r)
	}
}

func createLoadbalancer(service string) {

}
