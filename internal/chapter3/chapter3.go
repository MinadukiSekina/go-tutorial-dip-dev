package chapter3

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

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
	query := r.URL.Query()
	var names []string
	var ok bool
	names, ok = query["name"]
	if !ok {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}
	params := map[string][]string{
		"name": names,
	}

	// データ受信用のチャンネル
	ch1 := make(chan []int)
	// エラー受信用のチャンネル
	errch := make(chan error)

	// ユーザー情報を取得する
	go GetUserID(ctx, ch1, errch, params)

	var ids []int
	select {
	case ids = <-ch1:
		// ユーザーが見つからない場合はエラーを返す
		if len(ids) == 0 {
			http.Error(w, "User is not found", http.StatusNotFound)
			return
		}
	case err = <-errch:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// クエリパラメータ用のmapをリセット
	delete(params, "name")
	for _, id := range ids {
		params["userID"] = append(params["userID"], strconv.Itoa(id))
	}

	// データ受信用のチャンネル
	ch2 := make(chan []Entry)

	// データを取得する
	go GetEntries(ctx, ch2, errch, params)

	var entries []Entry
	select {
	case entries = <-ch2:
	case err = <-errch:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 値を返却する
	data := map[string][]Entry{"entries": entries}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Encoding json is failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

}

func GetUserID(ctx context.Context, ch chan []int, errch chan error, params map[string][]string) {
	// ヘッダーの設定
	header := map[string][]string{"key": {"dip"}}

	var err error
	// Clientのインスタンス化
	var c *networking.Client
	c, err = networking.NewClient(targetURL)
	if err != nil {
		errch <- err
		return
	}
	// 外部APIへリクエスト
	var res *http.Response
	res, err = c.NewRequestAndDo(ctx, http.MethodGet, c.BaseURL.JoinPath("/users"), header, params, nil)
	if err != nil {
		errch <- err
		return
	}
	defer res.Body.Close()

	got := []User{}
	if err = json.NewDecoder(res.Body).Decode(&got); err != nil {
		errch <- err
		return
	}

	var ids []int
	for _, user := range got {
		ids = append(ids, user.ID)
	}

	ch <- ids
}

func GetEntries(ctx context.Context, ch chan []Entry, errch chan error, params map[string][]string) {
	// ヘッダーの設定
	header := map[string][]string{"key": {"dip"}}

	var err error
	// Clientのインスタンス化
	var c *networking.Client
	c, err = networking.NewClient(targetURL)
	if err != nil {
		errch <- err
		return
	}

	// 外部APIへリクエスト
	var res *http.Response
	res, err = c.NewRequestAndDo(ctx, http.MethodGet, c.BaseURL.JoinPath("/entries"), header, params, nil)
	if err != nil {
		errch <- err
		return
	}
	defer res.Body.Close()

	var entries []Entry
	if err = json.NewDecoder(res.Body).Decode(&entries); err != nil {
		errch <- err
		return
	}

	ch <- entries
}