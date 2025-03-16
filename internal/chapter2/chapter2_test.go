package chapter2

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/dip-dev/go-tutorial/internal/helper/test"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TimeoutOption(timeout time.Duration) Option {
	return func(c *Client) {
		c.client.Timeout = timeout
	}
}

func TestNewClient(t *testing.T) {
	t.Run("正常ケース:オプションの設定あり", func(t *testing.T) {
		baseURL := "http://mock:80"
		timeout := 10 * time.Second
		opts := []Option{
			TimeoutOption(timeout),
		}
		client, err := NewClient(baseURL, opts...)
		assert.NoError(t, err)
		assert.Equal(t, timeout, client.client.Timeout)
	})
	t.Run("正常ケース:baseURLが正しい", func(t *testing.T) {
		baseURL := "http://mock:80"
		url, _ := url.Parse(baseURL)
		client, err := NewClient(baseURL)
		assert.NoError(t, err)
		assert.Equal(t, url, client.baseURL)
	})
	t.Run("異常ケース:baseURLが不正", func(t *testing.T) {
		oldBaseURL := os.Getenv("MOCK_API_URL")
		defer os.Setenv("MOCK_API_URL", oldBaseURL)

		baseURL := ":\\test"
		os.Setenv("MOCK_API_URL", baseURL)
		_, err := NewClient(baseURL)
		assert.Error(t, err)
	})
}

func TestNewRequestAndDo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	success := map[string]struct {
		method string
		url    url.URL
		header map[string][]string
		params map[string]string
		body   any
	}{
		"正常ケース:ヘッダー・パラメータあり": {
			method: http.MethodGet,
			url: url.URL{
				Scheme: "http",
				Host:   "mock:80",
			},
			header: map[string][]string{
				"key": {"dip"},
			},
			params: map[string]string{
				"age": "25",
			},
		},
		"正常ケース:ヘッダー・パラメータなし": {
			method: http.MethodGet,
			url: url.URL{
				Scheme: "http",
				Host:   "mock:80",
			},
		},
		"正常ケース:ヘッダー・パラメータなし・ボディがエンコード文字列": {
			method: http.MethodPost,
			url: url.URL{
				Scheme: "http",
				Host:   "mock:80",
			},
			body: url.Values{"name": {"dip 次郎"}, "age": {"24"}}.Encode(),
		},
		"正常ケース:ヘッダー・パラメータなし・ボディをJSONとして扱う": {
			method: http.MethodPost,
			url: url.URL{
				Scheme: "http",
				Host:   "mock:80",
			},
			body: map[string]string{"name": "dip 次郎", "age": "24"},
		},
	}
	fail := map[string]struct {
		method string
		url    url.URL
		header map[string][]string
		params map[string]string
		body   any
	}{
		"異常ケース:baseURLが不正": {
			method: http.MethodGet,
			url: url.URL{
				Scheme: "http",
				Host:   "invalid-url",
			},
		},
	}

	for tn, tc := range success {
		t.Run(tn, func(t *testing.T) {
			c, _ := NewClient(targetURL)
			res, err := c.NewRequestAndDo(ctx, tc.method, &tc.url, tc.header, tc.params, tc.body)
			if err != nil {
				t.Errorf("error: %#v", err)
				return
			}
			defer res.Body.Close()
		})
	}
	for tn, tc := range fail {
		t.Run(tn, func(t *testing.T) {
			c, _ := NewClient(targetURL)
			res, err := c.NewRequestAndDo(ctx, tc.method, &tc.url, tc.header, tc.params, tc.body)
			assert.Error(t, err)
			if res != nil {
				defer res.Body.Close()
			}
		})
	}
}

func TestWithHTTPClient(t *testing.T) {
	t.Run("正常ケース:HTTPClientが設定される", func(t *testing.T) {
		httpClient := &http.Client{}
		client, err := NewClient("http://mock:80", WithHTTPClient(httpClient))
		assert.NoError(t, err)
		assert.Equal(t, httpClient, client.client)
	})
}

func TestGet(t *testing.T) {
	success := map[string]struct {
		params     map[string]string
		response   []User
		wantStatus int
	}{
		"正常ケース": {
			params: map[string]string{
				"age": "25",
			},
			response: []User{
				{
					Name: "dip 太郎",
					Age:  25,
				},
				{
					Name: "dip 花子",
					Age:  25,
				},
			},
			wantStatus: http.StatusOK,
		},
	}
	fail := map[string]struct {
		method     string
		wantStatus int
	}{
		"異常: Getメソッドではない": {
			method:     http.MethodPost,
			wantStatus: http.StatusMethodNotAllowed,
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
			Get(w, r)
			got := []User{}
			t.Logf(w.Body.String())
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Errorf("errpr: %#v, res: %#v", err, got)
			}
			assert.Equal(t, tc.wantStatus, w.Code)
			assert.ElementsMatch(t, got, tc.response)
		})
	}
	for tn, tc := range fail {
		t.Run(tn, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tc.method, "http://localhost/", nil)
			Get(w, r)
			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
	t.Run("異常: パラメータのパース失敗", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "http://localhost/?%", nil)
		w := httptest.NewRecorder()

		Get(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("異常: ボディのコピーエラー", func(t *testing.T) {
		param := url.Values{}
		param.Add("age", "25")

		r := httptest.NewRequest(http.MethodGet, "http://localhost/?"+param.Encode(), nil)
		// エンコード処理でエラーを返すカスタムResponseWriterを利用
		errW := &test.ErrorResponseWriter{}

		// 呼び出し
		Get(errW, r)

		assert.Equal(t, http.StatusInternalServerError, errW.Code())
	})
}

func TestCreate(t *testing.T) {
	success := map[string]struct {
		params     map[string]string
		response   User
		wantStatus int
	}{
		"正常ケース": {
			params: map[string]string{
				"name": "dip 次郎",
				"age":  "24",
			},
			response: User{
				Name: "dip 次郎",
				Age:  24,
			},
			wantStatus: http.StatusOK,
		},
	}
	fail := map[string]struct {
		method     string
		params     map[string]string
		wantStatus int
	}{
		"異常ケース：Getメソッド": {
			method:     http.MethodGet,
			wantStatus: http.StatusMethodNotAllowed,
		},
		"異常ケース：パラメータが不正（nameが空）": {
			params:     map[string]string{"name": "", "age": "24"},
			wantStatus: http.StatusBadRequest,
		},
		"異常ケース：パラメータが不正（ageが空）": {
			params:     map[string]string{"name": "dip 次郎", "age": ""},
			wantStatus: http.StatusBadRequest,
		},
		"異常ケース：パラメータが不正（ageが文字列）": {
			params:     map[string]string{"name": "dip 次郎", "age": "twenty"},
			wantStatus: http.StatusBadRequest,
		},
		"異常ケース：パラメータが不正（すべて空）": {
			params:     map[string]string{"name": "", "age": ""},
			wantStatus: http.StatusBadRequest,
		},
		"異常ケース：パラメータが不正（name無し）": {
			params:     map[string]string{"age": "24"},
			wantStatus: http.StatusBadRequest,
		},
		"異常ケース：パラメータが不正（age無し）": {
			params:     map[string]string{"name": "dip 次郎"},
			wantStatus: http.StatusBadRequest,
		},
		"異常ケース：パラメータなし": {
			wantStatus: http.StatusBadRequest,
		},
	}

	for tn, tc := range success {
		t.Run(tn, func(t *testing.T) {
			params, err := json.Marshal(tc.params)
			if err != nil {
				t.Errorf("error: %#v", err)
				return
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "http://localhost/", bytes.NewReader(params))
			// Content-Typeヘッダーを追加
			r.Header.Set("Content-Type", "application/json")

			Create(w, r)

			var got User
			if err = json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Errorf("error: %#v, res: %#v", err, got)
			}
			assert.Equal(t, tc.wantStatus, w.Code)
			assert.Equal(t, tc.response, got)
		})
	}
	for tn, tc := range fail {
		t.Run(tn, func(t *testing.T) {
			params, err := json.Marshal(tc.params)
			if err != nil {
				t.Errorf("error: %#v", err)
				return
			}

			w := httptest.NewRecorder()
			var r *http.Request
			if tc.method == http.MethodGet {
				r = httptest.NewRequest(tc.method, "http://localhost/?", nil)
			} else {
				r = httptest.NewRequest(http.MethodPost, "http://localhost/", bytes.NewReader(params))
			}

			Create(w, r)

			assert.Equal(t, tc.wantStatus, w.Code)
		})
	}
	t.Run("異常: パラメータのJSONデコード失敗", func(t *testing.T) {
		params := "test"
		r := httptest.NewRequest(http.MethodPost, "http://localhost/", strings.NewReader(params))
		w := httptest.NewRecorder()

		// 呼び出し
		Create(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("異常: ボディのコピーエラー", func(t *testing.T) {
		params, _ := json.Marshal(map[string]string{
			"name": "dip 次郎",
			"age":  "24",
		})
		r := httptest.NewRequest(http.MethodPost, "http://localhost/", bytes.NewReader(params))
		// エンコード処理でエラーを返すカスタムResponseWriterを利用
		errW := &test.ErrorResponseWriter{}

		// 呼び出し
		Create(errW, r)

		assert.Equal(t, http.StatusInternalServerError, errW.Code())
	})
}
