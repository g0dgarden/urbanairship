package urbanairship

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"encoding/json"
)

// Push push APIにリクエストするJSONの構造体です
type Push struct {
	Audience     *Audience     `json:"audience"`
	Notification *Notification `json:"notification"`
	DeviceTypes  []string      `json:"device_types"`
}

// Notification 通知のメッセージとDeepLinkを設定します
type Notification struct {
	Alert   string `json:"alert"`
	Actions struct {
		Open struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		} `json:"open"`
	} `json:"actions"`
}

// Audience 通知対象のチャンネルIDを設定します
type Audience struct {
	IosChannel     []string `json:"ios_channel,omitempty"`
	AndroidChannel []string `json:"android_channel,omitempty"`
}

// PushResponse push APIのレスポンス形式です
type PushResponse struct {
	Ok          bool          `json:"ok"`
	OperationID string        `json:"operation_id"`
	PushIds     []string      `json:"push_ids"`
	MessageIds  []interface{} `json:"message_ids"`
	ContentUrls []interface{} `json:"content_urls"`
}

// NewPush はpush apiのリクエスト形式の構造体のコンストラクタです
func NewPush(audience *Audience, notification *Notification, deviceTypes []string) (*Push, error) {
	// check parameters
	if audience == nil {
		return nil, errors.New("missing audience")
	}
	if notification == nil {
		return nil, errors.New("missing notification")
	}
	if len(deviceTypes) == 0 {
		return nil, errors.New("missing device types")
	}

	p := &Push{
		Audience:     audience,
		Notification: notification,
		DeviceTypes:  deviceTypes,
	}
	return p, nil
}

// doPushRequest はUrbanAirshipのpushAPIへリクエストする実装です
// See: http://docs.urbanairship.com/api/ua.html#push
func (c *Client) doPushRequest(ctx context.Context, body *Push) (*http.Response, error) {
	endpoint := "/api/push"

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(ctx, "POST", endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	res, err := checkResponse(c.HTTPClient.Do(req))
	if err != nil {
		return nil, fmt.Errorf("faild urban push request. err:%v", err)
	}
	return res, nil
}

// Push はpushAPIのレスポンス結果を構造体にマッピングし返します。
func (c *Client) Push(ctx context.Context, body *Push) (*PushResponse, error) {

	// Check parameters
	if ctx == nil {
		return nil, errors.New("missing context")
	}

	if body == nil {
		return nil, errors.New("missing urban request body")
	}

	res, err := checkResponse(c.Urban.doPushRequest(ctx, body))
	if err != nil {
		return nil, err
	}

	var push PushResponse
	if err := decodeBody(res, &push); err != nil {
		return nil, err
	}
	return &push, nil

}
