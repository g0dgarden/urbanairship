package urbanairship

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckResponse(t *testing.T) {
	assert := assert.New(t)

	body :=
		ioutil.NopCloser(strings.NewReader(
			`{
			"error":"hoge",
			"error_code":40530,
			"details":{"error":"fuga"}
		}`))

	cases := []struct {
		inResponse  *http.Response
		inErr       error
		outResponse *http.Response
		outErr      error
	}{

		{inResponse: &http.Response{StatusCode: 200, Status: ""}, inErr: nil, outResponse: &http.Response{StatusCode: 200}, outErr: nil},
		{inResponse: &http.Response{StatusCode: 201, Status: ""}, inErr: nil, outResponse: &http.Response{StatusCode: 201}, outErr: nil},
		{inResponse: &http.Response{StatusCode: 202, Status: ""}, inErr: nil, outResponse: &http.Response{StatusCode: 202}, outErr: nil},
		{inResponse: &http.Response{StatusCode: 204, Status: ""}, inErr: nil, outResponse: &http.Response{StatusCode: 204}, outErr: nil},
		{inResponse: &http.Response{StatusCode: 400, Status: "", Body: body}, inErr: nil, outResponse: nil, outErr: errors.New("urban error response. error_code: 40530 error :hoge details_error :fuga")},
		{inResponse: &http.Response{StatusCode: 401, Status: ""}, inErr: nil, outResponse: nil, outErr: errors.New("authentication failed")},
		{inResponse: &http.Response{StatusCode: 404, Status: ""}, inErr: nil, outResponse: nil, outErr: errors.New("resource not found")},
		{inResponse: &http.Response{StatusCode: 500, Status: "エラーだよ"}, inErr: nil, outResponse: nil, outErr: errors.New("client: エラーだよ")},
	}

	for _, v := range cases {
		res, err := checkResponse(v.inResponse, v.inErr)

		if v.outResponse != nil {
			assert.Equal(res, v.outResponse)
		} else {
			assert.Equal((*http.Response)(nil), res)
		}

		if v.inResponse.StatusCode == 400 {
			assert.EqualError(err, v.outErr.Error())
		} else {
			assert.Equal(err, v.outErr)
		}
	}
}
