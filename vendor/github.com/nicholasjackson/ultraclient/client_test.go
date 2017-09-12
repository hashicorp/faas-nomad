package ultraclient

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var client *ClientImpl
var loadbalancingStrategy MockLoadbalancingStrategy
var backoffStrategy MockBackoffStrategy
var mockStats MockStats
var urls = []url.URL{url.URL{Host: "something:3232"}, url.URL{Host: "somethingelse:2323"}}
var urlIndex = 0

var getURL GetEndpoint = func() url.URL {
	url := urls[urlIndex]

	if urlIndex == 1 {
		urlIndex = 0
	} else {
		urlIndex = 1
	}

	return url
}

func setupClient(retryCount int) {
	urlIndex = 0

	loadbalancingStrategy = MockLoadbalancingStrategy{}
	loadbalancingStrategy.On("SetEndpoints", mock.Anything)
	loadbalancingStrategy.On("NextEndpoint").Return(getURL)
	loadbalancingStrategy.On("GetEndpoints").Return(urls)
	loadbalancingStrategy.On("Length").Return(len(urls))
	loadbalancingStrategy.On("Clone")

	var retries []time.Duration
	for i := 0; i < retryCount; i++ {
		retries = append(retries, 1*time.Millisecond)
	}

	backoffStrategy = MockBackoffStrategy{}
	backoffStrategy.On("Create", mock.Anything, mock.Anything).
		Return(retries)

	mockStats = MockStats{}
	mockStats.On("Timing", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockStats.On("Increment", mock.Anything, mock.Anything, mock.Anything)

	client = NewClient(
		Config{
			RetryDelay:             100 * time.Millisecond,
			Retries:                retryCount,
			Timeout:                10 * time.Millisecond,
			ErrorPercentThreshold:  50,
			DefaultVolumeThreshold: 2,
			StatsD: StatsD{
				Prefix: "myapp",
				Tags:   []string{"env:production"},
			},
		},
		&loadbalancingStrategy,
		&backoffStrategy,
	).(*ClientImpl)

	client.RegisterStats(&mockStats)

	hystrix.Flush()
}

func TestNewRailsSessionSetsRetriesToURLsLengthIfNotSet(t *testing.T) {
	setupClient(0)
	c := NewClient(
		Config{RetryDelay: 100 * time.Millisecond},
		&loadbalancingStrategy,
		&backoffStrategy,
	).(*ClientImpl)

	assert.Equal(t, 1, c.config.Retries)
}

func TestNewRailsSessionSetsRetriesIfSet(t *testing.T) {
	setupClient(0)
	c := NewClient(
		Config{Retries: 3, RetryDelay: 100 * time.Millisecond},
		&loadbalancingStrategy,
		&backoffStrategy,
	).(*ClientImpl)

	assert.Equal(t, 3, c.config.Retries)
}

func TestDoCallsCommand(t *testing.T) {
	setupClient(0)

	callCount := 0

	err := client.Do(func(endpoint url.URL) error {
		callCount++
		return nil
	})

	assert.Nil(t, err)
	assert.Equal(t, 1, callCount)
}

func TestClientCallsLoadBalancer(t *testing.T) {
	setupClient(0)

	err := client.Do(func(endpoint url.URL) error {
		return nil
	})

	assert.Nil(t, err)
	loadbalancingStrategy.AssertCalled(t, "NextEndpoint")
}

func TestClientCallIncrementsStats(t *testing.T) {
	setupClient(0)

	err := client.Do(func(endpoint url.URL) error {
		return nil
	})

	assert.Nil(t, err)

	tags := append(client.config.StatsD.Tags, "server:something_3232")
	mockStats.AssertCalled(t,
		"Increment",
		"myapp.called", tags, mock.Anything)
}

func TestClientCallTimingStats(t *testing.T) {
	setupClient(0)

	err := client.Do(func(endpoint url.URL) error {
		return nil
	})

	assert.Nil(t, err)

	tags := append(client.config.StatsD.Tags, "server:something_3232")
	mockStats.AssertCalled(t,
		"Timing",
		"myapp.timing", tags, mock.Anything, mock.Anything)
}

func TestClientRetriesWithDifferentURLAndReturnsError(t *testing.T) {
	setupClient(2)

	var urls []url.URL
	err := client.Do(func(endpoint url.URL) error {
		urls = append(urls, endpoint)
		return fmt.Errorf("aaah")
	})

	assert.NotNil(t, err)
}

func TestSuccessIncrementsStats(t *testing.T) {
	setupClient(0)
	err := client.Do(func(endpoint url.URL) error {
		return nil
	})

	assert.Nil(t, err)

	tags := append(client.config.StatsD.Tags, "server:something_3232")
	mockStats.AssertCalled(t,
		"Increment",
		"myapp.called", tags, mock.Anything)
}

func TestTimeoutReturnsError(t *testing.T) {
	setupClient(0)

	err := client.Do(func(endpoint url.URL) error {
		time.Sleep(150 * time.Millisecond)
		return nil
	})

	clientError := err.(ClientError)

	assert.Equal(t, ErrorTimeout, clientError.Message)
}

func TestTimeoutIncrementsStats(t *testing.T) {
	setupClient(0)

	err := client.Do(func(endpoint url.URL) error {
		time.Sleep(150 * time.Millisecond)
		return nil
	})

	assert.Equal(t, ErrorTimeout, err.(ClientError).Message)

	tags := append(client.config.StatsD.Tags, "server:something_3232")
	mockStats.AssertCalled(t,
		"Increment",
		"myapp.timeout", tags, mock.Anything)
}

func TestOpenCircuitReturnsError(t *testing.T) {
	setupClient(4)

	err := client.Do(func(endpoint url.URL) error {
		time.Sleep(150 * time.Millisecond)
		return nil
	})

	clientError := err.(ClientError)

	assert.Equal(t, ErrorCircuitOpen, clientError.Message)
}

func TestOpenCircuitIncrementsStats(t *testing.T) {
	setupClient(5)

	err := client.Do(func(endpoint url.URL) error {
		time.Sleep(150 * time.Millisecond)
		return nil
	})

	assert.Equal(t, ErrorCircuitOpen, err.(ClientError).Message)

	tags1 := append(client.config.StatsD.Tags, "server:something_3232")
	tags2 := append(client.config.StatsD.Tags, "server:somethingelse_2323")
	mockStats.AssertCalled(t,
		"Increment",
		"myapp.timeout", tags1, mock.Anything)
	mockStats.AssertCalled(t,
		"Increment",
		"myapp.timeout", tags2, mock.Anything)
	mockStats.AssertCalled(t,
		"Increment",
		"myapp.circuitopen", tags1, mock.Anything)
}

func TestCloneCreatesACloneOfTheClient(t *testing.T) {
	setupClient(0)
	c := client.Clone().(*ClientImpl)

	assert.NotEqual(t, client, c)
	loadbalancingStrategy.AssertCalled(t, "Clone")

	assert.Equal(t, client.config, c.config)
	assert.Equal(t, client.backoffStrategy, c.backoffStrategy)
	assert.Equal(t, client.statsCollection, c.statsCollection)
	assert.Equal(t, client.retry, c.retry)
}
