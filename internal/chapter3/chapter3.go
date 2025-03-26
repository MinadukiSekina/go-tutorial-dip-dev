package chapter3

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dip-dev/go-tutorial/internal/helper/networking"
)

type User struct {
	ID   int
	Name string
	Age  int
}

type Entry struct {
	Name   string
	UserID int
	Salary int
}

const targetURL = "http://mock-api"

func Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// クエリパラメータの設定
	var err error
	if err = r.ParseForm(); err != nil {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}
	params := map[string][]string{}
	for k, v := range r.Form {
		params[k] = append(params[k], v...)
	}

	_, err = GetUserID(ctx, params)

	if err != nil {
		http.Error(w, "Getting user is failed", http.StatusInternalServerError)
		return
	}
}

func GetUserID(ctx context.Context, params map[string][]string) (ids []int, err error) {
	// ヘッダーの設定
	header := map[string][]string{"key": {"dip"}}

	// Clientのインスタンス化
	var c *networking.Client
	c, err = networking.NewClient(targetURL)
	if err != nil {
		return ids, err
	}
	// 外部APIへリクエスト
	var res *http.Response
	res, err = c.NewRequestAndDo(ctx, http.MethodGet, c.BaseURL.JoinPath("/users"), header, params, nil)
	if err != nil {
		return ids, err
	}
	defer res.Body.Close()

	got := []User{}
	if err = json.NewDecoder(res.Body).Decode(&got); err != nil {
		return ids, err
	}

	for _, user := range got {
		ids = append(ids, user.ID)
	}

	return ids, nil
}