package handlers

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/hashicorp/faas-nomad/consul"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/ultraclient"
	cache "github.com/patrickmn/go-cache"
)

// MakeProxy creates a proxy for HTTP web requests which can be routed to a function.
func MakeProxy(client ProxyClient, resolver consul.ServiceResolver, statsDAddress string, logger hclog.Logger, stats *statsd.Client) http.HandlerFunc {
	c := cache.New(5*time.Minute, 10*time.Minute)
	p := &Proxy{
		lbCache:       c,
		client:        client,
		resolver:      resolver,
		stats:         stats,
		statsDAddress: statsDAddress,
		logger:        logger.Named("proxy_client"),
	}

	return func(rw http.ResponseWriter, r *http.Request) {
		p.ServeHTTP(rw, r)
	}
}

// Proxy is a http.Handler which implements the ability to call a downstream function
type Proxy struct {
	lbCache       *cache.Cache
	client        ProxyClient
	resolver      consul.ServiceResolver
	stats         *statsd.Client
	statsDAddress string
	logger        hclog.Logger
}

func (p *Proxy) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != "POST" {
		p.stats.Incr("proxy.badrequest", []string{}, 1)
		p.logger.Error("Bad request", "request", r)

		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	service := r.Context().Value(FunctionNameCTXKey).(string)

	urls, _ := p.resolver.Resolve(service)
	if len(urls) == 0 {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	reqBody, _ := ioutil.ReadAll(r.Body)
	reqHeaders := r.Header
	defer r.Body.Close()

	var respBody []byte
	var respHeaders http.Header
	var err error

	lb := p.getLoadbalancer(service, urls, p.statsDAddress)
	lb.Do(func(endpoint url.URL) error {

		// add the querystring from the request
		endpoint.RawQuery = r.URL.RawQuery

		respBody, respHeaders, err = p.client.CallAndReturnResponse(endpoint.String(), reqBody, reqHeaders)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		p.logger.Error("Internal server error", "error", err)

		return
	}

	for k, v := range respHeaders {
		if len(v) > 0 {
			rw.Header().Set(k, v[0])
		}
	}

	rw.Write(respBody)
}

func (p *Proxy) getLoadbalancer(service string, endpoints []string, statsAddress string) ultraclient.Client {
	urls := make([]url.URL, 0)
	for _, e := range endpoints {
		url, err := url.Parse(e)
		if err != nil {
			log.Println(err)
		} else {
			urls = append(urls, *url)
		}
	}

	if lb, ok := p.lbCache.Get(service); ok {
		l := lb.(ultraclient.Client)
		l.UpdateEndpoints(urls)
		return l
	}

	lb := createLoadbalancer(urls, statsAddress, p.logger, service)
	p.lbCache.Set(service, lb, cache.DefaultExpiration)

	return lb
}

func createLoadbalancer(endpoints []url.URL, statsDAddr string, logger hclog.Logger, service string) ultraclient.Client {
	log := logger.Named("load_balancer")
	lb := ultraclient.RoundRobinStrategy{}
	bs := ultraclient.ExponentialBackoff{}
	sd, err := ultraclient.NewDogStatsD(url.URL{Host: statsDAddr})
	if err != nil {
		log.Error("Error creating statsD connection", "error", err)
	}

	config := ultraclient.Config{
		Timeout:                30 * time.Second,
		MaxConcurrentRequests:  1500,
		ErrorPercentThreshold:  25,
		DefaultVolumeThreshold: 10,
		Retries:                5,
		RetryDelay:             2 * time.Second,
		Endpoints:              endpoints,
		StatsD: ultraclient.StatsD{
			Prefix: "faas.nomadd.function",
			Tags: []string{
				"job:" + service,
			},
		},
	}

	c := ultraclient.NewClient(config, &lb, &bs)
	c.RegisterStats(sd)

	return c
}
