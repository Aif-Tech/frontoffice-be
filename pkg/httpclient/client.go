package httpclient

import (
	"net/http"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type DefaultClient struct {
	client *http.Client
}

func NewDefaultClient(timeout time.Duration) *DefaultClient {
	return &DefaultClient{
		client: &http.Client{Timeout: timeout},
	}
}

func (d *DefaultClient) Do(req *http.Request) (*http.Response, error) {
	return d.client.Do(req)
}
