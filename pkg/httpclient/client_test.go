package httpclient

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultClient_Do(t *testing.T) {
	client := NewDefaultClient(5 * time.Second)

	req, err := http.NewRequest("GET", "https://example.com", nil)
	assert.NoError(t, err)

	resp, err := client.Do(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
