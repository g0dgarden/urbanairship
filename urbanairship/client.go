package urbanairship

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

const (
	apiVersion = "3"
	baseURL    = "https://go.urbanairship.com"
	mimeType   = "application/vnd.urbanairship+json; version=" + apiVersion
)

// Executor は UrbanAirship へのAPIへリクエストするインターフェースです。
// インターフェースとして定義することで、Mockの差し込みが可能となりテスタラブルな実装可能になります
type Executor interface {
	// doPushRequestは UrbanAirship.push APIへリクエストを行います
	doPushRequest(ctx context.Context, body *Push) (*http.Response, error)
}

// ErrResponse は UrbanAirship へのエラー時のHTTPレスポンスのフォーマットです。
type ErrResponse struct {
	Ok        bool   `json:"ok"`
	Message   string `json:"error"`
	ErrorCode int    `json:"error_code"`
	Details   struct {
		Message string `json:"error"`
	} `json:"details"`
	OperationID string `json:"operation_id"`
}

// Error インターフェースの実装
func (e *ErrResponse) Error() string {
	return fmt.Sprintf(
		"urban error response. error_code: %v error :%v details_error :%v",
		e.ErrorCode, e.Message, e.Details.Message)
}

// Client は UrbanAirship のAPIを扱うためのクライアント
type Client struct {
	Urban Executor
}

type UrbanAirship struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
	MimeType   string

	Username, Password string
}

// NewClient はUrbanAirshipへリクエストするClientを生成します。
func NewClient(user, pass string) (*Client, error) {

	if len(user) == 0 {
		return nil, errors.New("missing  username")
	}

	if len(pass) == 0 {
		return nil, errors.New("missing user password")
	}

	parsedURL, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, err
	}

	u := &UrbanAirship{
		BaseURL:    parsedURL,
		HTTPClient: http.DefaultClient,
		MimeType:   mimeType,
		Username:   user,
		Password:   pass,
	}
	c := &Client{Urban: u}
	return c, nil
}

func (u *UrbanAirship) newRequest(ctx context.Context, method, spath string, body io.Reader) (*http.Request, error) {
	uri := u.BaseURL
	uri.Path = path.Join(u.BaseURL.Path, spath)
	req, err := http.NewRequest(method, uri.String(), body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", u.MimeType)
	req.SetBasicAuth(u.Username, u.Password)
	return req, nil
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
		return nil, parseErr(resp) // errorのJsonをパースしてerrorとして返します。
	case 401:
		return nil, errors.New("authentication failed")
	case 404:
		return nil, errors.New("resource not found")
	default:
		return nil, fmt.Errorf("client: %s", resp.Status)
	}
}

// decodeBody はResponseBodyをoutに渡された、ポインタ型の構造体にマッピングします。
func decodeBody(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(out)
}

func parseErr(resp *http.Response) error {
	errResp := &ErrResponse{}
	if err := decodeBody(resp, errResp); err != nil {
		return fmt.Errorf("error decoding JSON body: %s", err)
	}
	return errResp
}
