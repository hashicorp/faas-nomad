package handlers

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

var postBody []byte

func testHandler(rw http.ResponseWriter, r *http.Request) {
	var err error
	postBody, err = ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	rw.Header().Add("TESTHeader", "somevalue")
	rw.Write([]byte("my body"))
}

func setupProxyClient(body []byte) (
	*HTTPProxyClient,
	*http.Request,
	*httptest.Server) {

	server := httptest.NewServer(http.HandlerFunc(
		testHandler,
	))

	r := httptest.NewRequest(
		"POST",
		"/system/function/testfunction",
		bytes.NewReader(body),
	)

	return MakeProxyClient(5*time.Second, hclog.Default()), r, server
}

func TestClientPostsGivenRequestBody(t *testing.T) {
	body := []byte("request body")
	c, r, s := setupProxyClient(body)
	defer s.Close()

	c.CallAndReturnResponse(s.URL, body, r.Header)

	assert.Equal(t, "request body", string(postBody))
}

func TestClientReturnsHeadersFromRequest(t *testing.T) {
	body := []byte("request body")
	c, r, s := setupProxyClient(body)
	defer s.Close()

	_, h, _ := c.CallAndReturnResponse(s.URL, body, r.Header)

	assert.Equal(t, "somevalue", h.Get("TESTHeader"))
}

func TestClientReturnsBodysFromRequest(t *testing.T) {
	body := []byte("request body")
	c, r, s := setupProxyClient(body)
	defer s.Close()

	b, _, _ := c.CallAndReturnResponse(s.URL, body, r.Header)

	assert.Equal(t, "my body", string(b))
}
