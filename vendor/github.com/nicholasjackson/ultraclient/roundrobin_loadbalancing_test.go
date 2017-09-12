package ultraclient

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

var rrStrategy RoundRobinStrategy
var endpoints = []url.URL{
	url.URL{Host: "host1"},
	url.URL{Host: "host2"},
}

func setupRRLB() {
	rrStrategy = RoundRobinStrategy{}
	rrStrategy.SetEndpoints(endpoints)
}

func TestFirstEndpointIsRandom(t *testing.T) {
	var urls []url.URL

	for i := 0; i < 10; i++ {
		setupRRLB()
		urls = append(urls, rrStrategy.NextEndpoint())
	}

	distinct := make(map[url.URL]int)
	for _, url := range urls {
		if distinct[url] == 0 {
			distinct[url]++
		} else {
			distinct[url] = 0
		}
	}

	assert.True(t, len(distinct) > 1)
}

func TestNextEndpointIsNextInArray(t *testing.T) {
	setupRRLB()

	first := rrStrategy.NextEndpoint()
	second := rrStrategy.NextEndpoint()

	assert.NotEqual(t, first, second)
}

func TestNextEndpointLoopsArray(t *testing.T) {
	setupRRLB()

	first := rrStrategy.NextEndpoint()
	_ = rrStrategy.NextEndpoint()
	third := rrStrategy.NextEndpoint()

	assert.Equal(t, first, third)
}

func TestCloneCreatesNewStrategy(t *testing.T) {
	setupRRLB()

	newStrategy := rrStrategy.Clone()

	assert.NotEqual(t, fmt.Sprintf("%p", &rrStrategy), fmt.Sprintf("%p", &newStrategy))
}
