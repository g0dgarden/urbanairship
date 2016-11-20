package urbanairship

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// UrbanAirshipのインターフェースをみたすdummyのstruct
type fakeClient struct {
	// インターフェース埋め込み
	UrbanAirship
	FakeDoPushRequest func(ctx context.Context, body *Push) (*http.Response, error)
}

// doPushRequestのdummyの実装。偽装したダミーのレスポンスを返します。偽装することで、実際にHTTP通信をさせないでのテストが可能となります。
func (c *fakeClient) doPushRequest(ctx context.Context, body *Push) (*http.Response, error) {
	return c.FakeDoPushRequest(ctx, body)
}

// push api からのレスポンスが成功し、構造体に正しくマッピングし取得出来ているか確認します。
func TestSuccessPush(t *testing.T) {
	assert := assert.New(t)

	// mockの関数を差し込みレスポンスを偽装します。
	fakeCli := &fakeClient{
		FakeDoPushRequest: func(_ context.Context, _ *Push) (*http.Response, error) {
			// 成功時のレスポンス
			body :=
				`{
				"ok":true,"operation_id":"ef77412e-c336-4575-b480-b9ca33bb6964",
				"push_ids":["5b9b1152-a4ac-4fa3-9e90-b699ac286de2"],
				"message_ids":[],
				"content_urls":[]
			}`
			res := &http.Response{
				StatusCode: 202,
				Body:       ioutil.NopCloser(strings.NewReader(body)),
			}
			return res, nil
		},
	}

	expected := &PushResponse{
		Ok:      true,
		PushIds: []string{"5b9b1152-a4ac-4fa3-9e90-b699ac286de2"},
	}

	client := Client{Urban: fakeCli}
	resp, err := client.Push(context.Background(), &Push{})

	assert.Nil(err)
	assert.Equal(expected.Ok, resp.Ok)
	assert.Equal(expected.PushIds, resp.PushIds)
}

// push api からのレスポンスが 202以外の時に正しくエラーハンドリング出来ているか確認します。
func TestBadRequestForPush(t *testing.T) {
	assert := assert.New(t)

	// ----------------------------------------
	// status code が202以外で返ってくるケース
	// ----------------------------------------
	fakeCli := &fakeClient{
		FakeDoPushRequest: func(_ context.Context, _ *Push) (*http.Response, error) {
			// 失敗時のレスポンス
			body :=
				`{
				"ok":false,
				"error":"Could not parse request body.",
				"error_code":40000,
				"details":{"error":"The key '' is not allowed in this context",
				"path":"notification.",
				"location":{"line":1,"column":46}},"operation_id":"dff10f45-685b-4b90-a719-7744bd24bc8f"
			}`

			res := &http.Response{
				StatusCode: 400,
				Body:       ioutil.NopCloser(strings.NewReader(body)),
			}
			return res, nil
		},
	}

	client := Client{Urban: fakeCli}

	expected := "status code 400 is not 202"
	resp, err := client.Push(context.Background(), &Push{})

	// ----------------------------------------
	// errorが返ってくるケース
	// ----------------------------------------
	fakeCli = &fakeClient{
		FakeDoPushRequest: func(ctx context.Context, body *Push) (*http.Response, error) {
			res := &http.Response{
				StatusCode: 400,
			}
			return res, errors.New("エラーだよー")
		},
	}

	client.Urban = fakeCli

	expected = "エラーだよー"
	resp, err = client.Push(context.Background(), &Push{})

	assert.Nil(resp)
	assert.Equal(expected, err.Error())

}

// 必須パラメーターをチェックするテストケースです。未設定の場合、対応したerror が返ってくることを確認します。
func TestRequiredParametersForPush(t *testing.T) {
	assert := assert.New(t)
	client := Client{}

	cases := []struct {
		ctx      context.Context
		body     *Push
		expected string
	}{
		// context がない
		{ctx: nil, body: &Push{}, expected: "missing context"},
		// body がない
		{ctx: context.Background(), body: nil, expected: "missing urban request body"},
	}

	for _, v := range cases {
		_, actual := client.Push(v.ctx, v.body)
		assert.Equal(v.expected, actual.Error())
	}
}
