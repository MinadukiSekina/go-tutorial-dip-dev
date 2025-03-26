package networking

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// 外部APIへリクエストするためのクライアント
type Client struct {
	BaseURL *url.URL
	Client  *http.Client
}

// クライアントの初期化処理
func NewClient(baseURL string, options ...Option) (*Client, error) {
	if envURL := os.Getenv("MOCK_API_URL"); envURL != "" {
		baseURL = envURL
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	c := &Client{
		BaseURL: u,
		Client:  &http.Client{},
	}
	for _, option := range options {
		option(c)
	}
	return c, nil
}

// API呼び出し時のオプション
type Option func(c *Client)

// Clientを上書きするオプション
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.Client = httpClient
	}
}

// リクエストの生成と実行
func (c *Client) NewRequestAndDo(ctx context.Context, method string, apiURL *url.URL, header map[string][]string, params map[string][]string, body any) (*http.Response, error) {
	var reqBody io.Reader

	switch v := body.(type) {
	case string:
		// form-urlencoded形式の文字列の場合
		reqBody = strings.NewReader(v)
	case nil:
		reqBody = nil
	default:
		// JSONの場合
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	// リクエストの作成
	req, err := http.NewRequestWithContext(ctx, method, apiURL.String(), reqBody)
	if err != nil {
		return nil, err
	}
	// ヘッダーの設定
	if header != nil {
		for k, vs := range header {
			for _, v := range vs {
				req.Header.Set(k, v)
			}
		}
	} else {
		req.Header.Set("Content-Type", "application/json")
	}
	// クエリパラメータの設定
	if params != nil {
		values := url.Values{}
		for k, param := range params {
			for _, v := range param {
				values.Add(k, v)
			}
		}
		req.URL.RawQuery = values.Encode()
	}
	// リクエストの実行
	return c.Client.Do(req)
}
