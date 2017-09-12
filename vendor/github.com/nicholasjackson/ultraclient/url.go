package ultraclient

import (
	"fmt"
	"net/url"
	"strings"
)

// PrettyPrintURL is a helper function to pretty print a url in a format
// suitable for statsd
func PrettyPrintURL(url *url.URL) string {
	parts := strings.Split(url.Host, ":")

	if len(parts) < 2 {
		return parts[0]
	}

	return fmt.Sprintf("%v_%v", parts[0], parts[1])
}
