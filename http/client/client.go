// Derived from reqfleet/shibuya; modified for lib.
package client

import (
	"net/http"
	"time"
)

var (
	ShortTimeout = 3 * time.Second
)

func MakeClientWithShortTimeout() *http.Client {
	return &http.Client{
		Timeout: ShortTimeout,
	}
}

func MakeClientWithCustomTimeout(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
	}
}
