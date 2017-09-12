package ultraclient

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetEndpointsSetsEndpoints(t *testing.T) {
	rs := RandomStrategy{}
	rs.SetEndpoints([]url.URL{url.URL{Host: "http://myhost.com"}})

	assert.Equal(t, "http://myhost.com", rs.endpoints[0].Host)
}

func TestSetEndpointsGetsRandomEndpoints(t *testing.T) {
	rs := RandomStrategy{}
	rs.SetEndpoints([]url.URL{
		url.URL{Host: "http://www1.myhost.com"},
		url.URL{Host: "http://www2.myhost.com"},
	})

	endpoint := rs.NextEndpoint()
	nextEndpoint := endpoint

	for i := 0; i < 100; i++ {
		nextEndpoint = rs.NextEndpoint()
		if nextEndpoint != endpoint {
			return
		}

		endpoint = nextEndpoint
	}

	t.Fatal("Should have returned randomised list of endpoints")
}

func TestCloneReturnsAnewInstance(t *testing.T) {
	rs := &RandomStrategy{}
	clone := rs.Clone()

	assert.NotNil(t, clone)
	assert.NotEqual(t, rs, clone)
}
