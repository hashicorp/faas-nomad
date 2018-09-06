package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/faas-nomad/consul"
	"github.com/hashicorp/faas-nomad/metrics"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/openfaas/faas-provider/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const infoTestVersion = "test"

func setupInfo(body string) (http.HandlerFunc, *httptest.ResponseRecorder, *http.Request) {

	mockStats := &metrics.MockStatsD{}
	mockStats.On("Incr", mock.Anything, mock.Anything, mock.Anything)
	mockStats.On("Gauge", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	mockServiceResolver = &consul.MockResolver{}
	mockServiceResolver.On("RemoveCacheItem", mock.Anything)

	logger := hclog.Default()

	return MakeInfo(logger, mockStats, infoTestVersion),
		httptest.NewRecorder(),
		httptest.NewRequest("GET", "/system/info", bytes.NewReader([]byte(body)))
}

func TestInfoReportsNomadProvider(t *testing.T) {

	h, rw, r := setupInfo("")

	h(rw, r)

	assert.Equal(t, http.StatusOK, rw.Code)
	infoRequest := types.InfoRequest{}

	body, err := ioutil.ReadAll(rw.Body)
	if err != nil {
		t.Fatal(err)
	}

	unmarshalErr := json.Unmarshal(body, &infoRequest)

	assert.Nil(t, unmarshalErr, "Expected no error")

	assert.Equal(t, infoRequest.Orchestration, "nomad")

	assert.Equal(t, infoRequest.Version.Release, infoTestVersion)
}
