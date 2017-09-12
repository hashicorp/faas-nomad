package handlers

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

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
	*httptest.ResponseRecorder,
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

	rw := httptest.NewRecorder()

	return MakeProxyClient(), rw, r, server
}

func TestClientPostsGivenRequestBody(t *testing.T) {
	body := []byte("request body")
	c, rw, r, s := setupProxyClient(body)
	defer s.Close()

	c.CallAndReturnResponse(s.URL, rw, r)

	assert.Equal(t, "request body", string(postBody))
}

func TestClientReturnsHeadersFromRequest(t *testing.T) {
	body := []byte("request body")
	c, rw, r, s := setupProxyClient(body)
	defer s.Close()

	c.CallAndReturnResponse(s.URL, rw, r)

	assert.Equal(t, "somevalue", rw.Header().Get("TESTHeader"))
}

func TestClientReturnsBodysFromRequest(t *testing.T) {
	body := []byte("request body")
	c, rw, r, s := setupProxyClient(body)
	defer s.Close()

	c.CallAndReturnResponse(s.URL, rw, r)

	assert.Equal(t, "my body", rw.Body.String())
}
