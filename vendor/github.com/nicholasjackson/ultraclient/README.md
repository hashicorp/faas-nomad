# UltraClient
Ultra client is a wrapper around exisiting packages to provide loadbalancing, circuit breaking and backoffs for networking code in go.  Rather than extend net/http you pass a function to ultraclient which accepts a url.URL, this allows you to use RPC based clients as well as net/http.

[![GoDoc](https://godoc.org/github.com/nicholasjackson/ultraclient?status.svg)](https://godoc.org/github.com/nicholasjackson/ultraclient)

[![CircleCI](https://circleci.com/gh/nicholasjackson/ultraclient.svg?style=svg)](https://circleci.com/gh/nicholasjackson/ultraclient)

## Usage

```go
lb := ultraclient.RoundRobinStrategy{}
bs := ultraclient.ExponentialBackoff{}
stats, _ := ultraclient.NewDogStatsD(url.URL{Host:"statsd:8125"})
endpoints := []url.URL{
  url.URL{Host: "server1:8080"},
  url.URL{Host: "server2:8080"},
}

config := loadbalancer.Config{
	Timeout:                50 * time.Millisecond,
	MaxConcurrentRequests:  500,
	ErrorPercentThreshold:  25,
	DefaultVolumeThreshold: 10,
  Retries:                100*time.Millisecond,
  Endpoints: endpoints,
	StatsD: loadbalancer.StatsD{
		Prefix: "application.client",
	},
}

client := ultraclient.NewClient(client, &lb, &bs)
client.RegisterStats(stats)
```

