package urbanairship

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/pkg/errors"
)

const (
	apiVersion = "3"
	baseURL    = "https://go.urbanairship.com"
	mimeType   = "application/vnd.urbanairship+json; version=" + apiVersion
)

// UrbanAirship はurbanへのAPIへリクエストするインターフェースです。
type UrbanAirship interface {
	doPushRequest(ctx context.Context, body *Push) (*http.Response, error)
}

// Client は UrbanAirship のAPIを扱うためのクライアント実装
type Client struct {
	Urban              UrbanAirship
	BaseURL            *url.URL
	HTTPClient         *http.Client
	Username, Password string
	MimeType           string
}

// NewClient はUrbanAirshipへリクエストするClientを生成します。
func NewClient(client *http.Client) (*Client, error) {
	if client == nil {
		client = http.DefaultClient
	}
	parsedURL, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, err
	}
	c := Client{
		BaseURL:    parsedURL,
		HTTPClient: client,
		MimeType:   mimeType,
	}
	return &c, nil
}

func (c *Client) newRequest(ctx context.Context, method, spath string, body io.Reader) (*http.Request, error) {
	u := *c.BaseURL
	u.Path = path.Join(c.BaseURL.Path, spath)
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", c.MimeType)
	req.SetBasicAuth(c.Username, c.Password)
	return req, nil
}

// ResponseBodyをoutに渡された、ポインタ型の構造体にマッピングします。
func decodeBody(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(out)
}

func checkResponse(resp *http.Response, err error) (*http.Response, error) {
	if err != nil {
		return resp, err
	}

	switch resp.StatusCode {
	case 200:
		return resp, nil
	case 201:
		return resp, nil
	case 202:
		return resp, nil
	case 204:
		return resp, nil
	case 400:
		return nil, errors.New("bad request") // TODO: error時のbodyをキャッチする
	case 401:
		return nil, errors.New("authentication failed")
	case 404:
		return nil, errors.New("resource not found")
	default:
		return nil, fmt.Errorf("client: %s", resp.Status)
	}
}
