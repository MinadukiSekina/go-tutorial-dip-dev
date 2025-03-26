package networking

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const targetURL = "http://mock-api"

func TimeoutOption(timeout time.Duration) Option {
	return func(c *Client) {
		c.Client.Timeout = timeout
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
		assert.Equal(t, timeout, client.Client.Timeout)
	})
	t.Run("正常ケース:baseURLが正しい", func(t *testing.T) {
		baseURL := "http://mock:80"
		url, _ := url.Parse(baseURL)
		client, err := NewClient(baseURL)
		assert.NoError(t, err)
		assert.Equal(t, url, client.BaseURL)
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
		params map[string][]string
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
			params: map[string][]string{
				"age": {"25"},
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
		params map[string][]string
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
	t.Run("異常ケース:JSONのマーシャルに失敗", func(t *testing.T) {
		c, _ := NewClient(targetURL)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// JSONにマーシャルできない値を渡す
		invalidBody := make(chan int)

		res, err := c.NewRequestAndDo(
			ctx,
			http.MethodPost,
			&url.URL{Scheme: "http", Host: "example.com"},
			nil,
			nil,
			invalidBody,
		)
		assert.Error(t, err)
		if res != nil {
			defer res.Body.Close()
		}
	})
	t.Run("異常ケース:リクエスト作成に失敗", func(t *testing.T) {
		c, _ := NewClient(targetURL)
		// 不正なメソッドを指定
		res, err := c.NewRequestAndDo(
			ctx,
			"\n", // 不正なHTTPメソッド
			&url.URL{Scheme: "http", Host: "example.com"},
			nil,
			nil,
			nil,
		)
		assert.Error(t, err)
		if res != nil {
			defer res.Body.Close()
		}
	})
}

func TestWithHTTPClient(t *testing.T) {
	t.Run("正常ケース:HTTPClientが設定される", func(t *testing.T) {
		httpClient := &http.Client{}
		client, err := NewClient("http://mock:80", WithHTTPClient(httpClient))
		assert.NoError(t, err)
		assert.Equal(t, httpClient, client.Client)
	})
}
