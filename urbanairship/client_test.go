package urbanairship

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	assert := assert.New(t)

	baseURL, _ := url.ParseRequestURI("https://go.urbanairship.com")
	expected := Client{
		BaseURL:    baseURL,
		HTTPClient: http.DefaultClient,
		MimeType:   "application/vnd.urbanairship+json; version=3",
	}

	client, err := NewClient(nil)
	assert.Nil(err)

	assert.Equal(expected.BaseURL, client.BaseURL)
	assert.Equal(expected.HTTPClient, client.HTTPClient)
	assert.Equal(expected.MimeType, client.MimeType)

}
