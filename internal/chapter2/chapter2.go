package chapter2

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dip-dev/go-tutorial/internal/helper/networking"
)

const targetURL = "http://mock-api"

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
	if name, ok := params["name"]; !ok || name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if age, ok := params["age"]; !ok || age == "" {
		http.Error(w, "age is required", http.StatusBadRequest)
		return
	}
	if _, err := strconv.Atoi(params["age"]); err != nil {
		http.Error(w, "age is not a number", http.StatusBadRequest)
		return
	}

	// フォームデータの作成
	formData := url.Values{}
	formData.Set("name", params["name"])
	formData.Set("age", params["age"])

	// Clientのインスタンス化
	c, err := networking.NewClient(targetURL)
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
		c.BaseURL.JoinPath("/users"),
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
	params := map[string][]string{}
	for k, v := range r.Form {
		params[k] = append(params[k], v...)
	}

	// Clientのインスタンス化
	c, err := networking.NewClient(targetURL)
	if err != nil {
		http.Error(w, "Failed to create client", http.StatusInternalServerError)
		return
	}
	// 外部APIへリクエスト
	res, err2 := c.NewRequestAndDo(ctx, http.MethodGet, c.BaseURL.JoinPath("/users"), header, params, nil)
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
