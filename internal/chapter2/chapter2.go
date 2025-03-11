package chapter2

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

const targetURL = "http://mock-api"

// 外部APIへリクエストするためのクライアント
type Client struct {
	baseURL *url.URL
	client  *http.Client
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
		baseURL: u,
		client:  &http.Client{},
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
		c.client = httpClient
	}
}

// リクエストの生成と実行
func (c *Client) NewRequestAndDo(ctx context.Context, method string, apiURL *url.URL, header map[string][]string, params map[string]string, body any) (*http.Response, error) {
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
		for k, v := range params {
			values.Add(k, v)
		}
		req.URL.RawQuery = values.Encode()
	}
	// リクエストの実行
	return c.client.Do(req)
}

type User struct {
	Name string
	Age  int
}

func Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// リクエストボディの設定
	var params map[string]string
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 必須パラメータのチェック
	if _, ok := params["name"]; !ok {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if _, ok := params["age"]; !ok {
		http.Error(w, "age is required", http.StatusBadRequest)
		return
	}

	// フォームデータの作成
	formData := url.Values{}
	formData.Set("name", params["name"])
	formData.Set("age", params["age"])

	// Clientのインスタンス化
	c, err := NewClient(targetURL)
	if err != nil {
		http.Error(w, "Failed to create client", http.StatusInternalServerError)
		return
	}

	// ヘッダーの設定
	header := map[string][]string{
		"key":          {"dip"},
		"Content-Type": {"application/x-www-form-urlencoded"},
	}

	// 外部APIへリクエスト
	res, err2 := c.NewRequestAndDo(
		ctx,
		http.MethodPost,
		c.baseURL.JoinPath("/users"),
		header,
		nil,
		formData.Encode(),
	)
	if err2 != nil {
		http.Error(w, err2.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	// ヘッダーをコピー
	for k, vs := range res.Header {
		for _, v := range vs {
			w.Header().Set(k, v)
		}
	}
	// ステータスコードをコピー
	w.WriteHeader(res.StatusCode)
	// ボディをコピー
	if _, err3 := io.Copy(w, res.Body); err3 != nil {
		http.Error(w, "Failed to copy body", http.StatusInternalServerError)
		return
	}
}

func Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ヘッダーの設定
	header := map[string][]string{"key": {"dip"}}
	// クエリパラメータの設定
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}
	var params = map[string]string{}
	for k, v := range r.Form {
		params[k] = v[0]
	}

	// Clientのインスタンス化
	c, err := NewClient(targetURL)
	if err != nil {
		http.Error(w, "Failed to create client", http.StatusInternalServerError)
		return
	}
	// 外部APIへリクエスト
	res, err2 := c.NewRequestAndDo(ctx, http.MethodGet, c.baseURL.JoinPath("/users"), header, params, nil)
	if err2 != nil {
		http.Error(w, err2.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	// ヘッダーをコピー
	for k, vs := range res.Header {
		for _, v := range vs {
			w.Header().Set(k, v)
		}
	}
	// ステータスコードをコピー
	w.WriteHeader(res.StatusCode)
	// ボディをコピー
	if _, err3 := io.Copy(w, res.Body); err3 != nil {
		http.Error(w, "Failed to copy body", http.StatusInternalServerError)
	}
}
