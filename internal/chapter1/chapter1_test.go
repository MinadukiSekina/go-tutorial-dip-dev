package chapter1

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dip-dev/go-tutorial/internal/helper/test"
)

func TestGetEcho(t *testing.T) {
	success := map[string]struct {
		params     map[string]string
		wantStatus int
	}{
		"正常: パラメータあり": {
			params: map[string]string{
				"name": "dip 太郎",
				"age":  "26",
			},
			wantStatus: http.StatusOK,
		},
		"正常: パラメータなし": {
			params:     map[string]string{},
			wantStatus: http.StatusOK,
		},
	}

	fail := map[string]struct {
		method     string
		params     map[string]string
		rawString  string
		wantStatus int
	}{
		"異常: Getメソッドではない": {
			method:     http.MethodPost,
			params:     map[string]string{},
			wantStatus: http.StatusMethodNotAllowed,
		},
		"異常: パラメータが不正": {
			method:     http.MethodGet,
			rawString:  "%",
			wantStatus: http.StatusBadRequest,
		},
	}

	for tn, tc := range success {
		t.Run(tn, func(t *testing.T) {
			param := url.Values{}
			for k, v := range tc.params {
				param.Add(k, v)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "http://localhost/?"+param.Encode(), nil)
			GetEcho(w, r)
			got := w.Body.String()
			assert.Equal(t, tc.wantStatus, w.Code)
			for k, v := range tc.params {
				assert.Contains(t, got, k)
				assert.Contains(t, got, v)
			}
		})
	}
	for tn, tc := range fail {
		t.Run(tn, func(t *testing.T) {
			form := url.Values{}
			for k, v := range tc.params {
				form.Add(k, v)
			}
			w := httptest.NewRecorder()
			var r *http.Request
			if tc.method == http.MethodGet {
				r = httptest.NewRequest(http.MethodGet, "http://localhost/?"+form.Encode()+tc.rawString, nil)
			} else {
				r = httptest.NewRequest(tc.method, "http://localhost/", strings.NewReader(form.Encode()))
			}
			GetEcho(w, r)
			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
	t.Run("異常: JSONエンコード失敗", func(t *testing.T) {
		// 正常なGETリクエストを作成
		r := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
		// エンコード処理でエラーを返すカスタムResponseWriterを利用
		errW := &test.ErrorResponseWriter{}

		// 呼び出し
		GetEcho(errW, r)

		assert.Equal(t, http.StatusInternalServerError, errW.Code())
	})
}
