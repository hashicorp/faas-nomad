package ultraclient

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrettyPrintsURLWithoutPort(t *testing.T) {
	u, _ := url.Parse("http://localhost/dfsdfsdf")

	s := PrettyPrintURL(u)

	assert.Equal(t, "localhost", s)
}

func TestPrettyPrintsURLWithPort(t *testing.T) {
	u, _ := url.Parse("http://localhost:3232/dfsdfsdf")

	s := PrettyPrintURL(u)

	assert.Equal(t, "localhost_3232", s)
}
