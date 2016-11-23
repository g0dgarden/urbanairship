package urbanairship

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Executorのインターフェースをみたすdummyのstruct
type dummyClient struct {
	FakeDoPushRequest func(ctx context.Context, body *Push) (*http.Response, error)
}

// doPushRequestのdummyの実装。偽装したダミーのレスポンスを返します。偽装することで、実際にHTTP通信をさせないでのテストが可能となります。
func (c *dummyClient) doPushRequest(ctx context.Context, body *Push) (*http.Response, error) {
	return c.FakeDoPushRequest(ctx, body)
}

// push api からのレスポンスが成功し、構造体に正しくマッピングし取得出来ているか確認します。
func TestSuccessPush(t *testing.T) {
	assert := assert.New(t)

	// mockの関数を差し込みレスポンスを偽装します。
	dummyCli := &dummyClient{
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

	client := Client{Urban: dummyCli}
	resp, err := client.Push(context.Background(), &Push{})

	assert.Nil(err)
	assert.Equal(expected.Ok, resp.Ok)
	assert.Equal(expected.PushIds, resp.PushIds)
}

// push api からのレスポンスが 202以外の時に正しくエラーハンドリング出来ているか確認します。
func TestRequestFailedForPush(t *testing.T) {
	assert := assert.New(t)

	// ----------------------------------------
	// status code:401(Unauthorized	) で errorが返ってくるケース
	// ----------------------------------------
	dummyCli := &dummyClient{
		FakeDoPushRequest: func(_ context.Context, _ *Push) (*http.Response, error) {
			// 失敗時のレスポンス
			body :=
				`{
				"ok":false,
				"error":"Could not parse request body.",
				"error_code":40203,
				"details":{"error":"'device_types' must be set"},
				"operation_id":"a2a7f724-a301-49ac-ab26-e46b0252d582"
			}`

			res := &http.Response{
				StatusCode: 401,
				Body:       ioutil.NopCloser(strings.NewReader(body)),
			}
			return res, nil
		},
	}

	client := Client{Urban: dummyCli}

	expected := "authentication failed"
	resp, err := client.Push(context.Background(), &Push{})

	assert.Nil(resp)
	assert.Equal(expected, err.Error())

	// ----------------------------------------
	// status code:400(bad request)で errorJsonがパースされて返ってくるケース
	// ----------------------------------------
	dummyCli = &dummyClient{
		FakeDoPushRequest: func(_ context.Context, _ *Push) (*http.Response, error) {
			body :=
				`{
				"ok":false,
				"error":"Could not parse request body.",
				"error_code":40530,
				"details":{"error":"Unrecognized device type 'xxx'","path":"device_types[0]","location":{"line":1,"column":199}},
				"operation_id":"7d312199-a40f-4a7d-9d25-ecb9ec24e917"}
			`

			res := &http.Response{
				StatusCode: 400,
				Body:       ioutil.NopCloser(strings.NewReader(body)),
			}
			return res, parseErr(res)
		},
	}

	client.Urban = dummyCli

	expected = "urban error response. error_code: 40530 error :Could not parse request body. details_error :Unrecognized device type 'xxx'"
	resp, err = client.Push(context.Background(), &Push{})

	assert.Nil(resp)
	assert.Equal(expected, err.Error())

}

// 必須パラメーターをチェックするテストケースです。
// 未設定の場合、対応したerror が返ってくることを確認します。
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
