package chapter3

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dip-dev/go-tutorial/internal/helper/test"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestGet(t *testing.T) {
	// レスポンスに含まれるキーの名前
	keyString := "entries"

	// 外部APIのモックサーバー用
	successHandlers := []test.Handler{
		{
			Path:    "/users",
			Handler: MockGetUser,
		},
		{
			Path:    "/entries",
			Handler: MockGetEntry,
		},
	}
	getUsersFailHandlers := []test.Handler{
		{
			Path: "/users",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/entries", http.StatusFound)
			},
		},
		{
			Path:    "/entries",
			Handler: MockGetEntry,
		},
	}
	getEntriesFailHandlers := []test.Handler{
		{
			Path:    "/users",
			Handler: MockGetUser,
		},
		{
			Path: "/entries",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/users", http.StatusFound)
			},
		},
	}

	success := map[string]struct {
		params     map[string][]string
		response   []Entry
		handlers   []test.Handler
		wantStatus int
	}{
		"正常ケース：データあり1": {
			params: map[string][]string{
				"name": {"dip 太郎"},
			},
			response: []Entry{
				{
					Name:   "案件情報1",
					UserID: 123456,
					Salary: 123456,
				},
			},
			handlers:   successHandlers,
			wantStatus: http.StatusOK,
		},
		"正常ケース：データあり2": {
			params: map[string][]string{
				"name": {"dip 次郎"},
			},
			response: []Entry{
				{
					Name:   "案件情報2",
					UserID: 234567,
					Salary: 123456,
				},
			},
			handlers:   successHandlers,
			wantStatus: http.StatusOK,
		},
	}
	fail := map[string]struct {
		method     string
		params     map[string][]string
		resWriter  http.ResponseWriter
		handlers   []test.Handler
		wantStatus int
	}{
		"異常ケース：Getメソッドではない": {
			method: http.MethodPost,
			params: map[string][]string{
				"name": {"dip 太郎"},
			},
			handlers:   successHandlers,
			wantStatus: http.StatusMethodNotAllowed,
		},
		"異常ケース：ユーザーデータなし": {
			method: http.MethodGet,
			params: map[string][]string{
				"name": {"dip 三郎"},
			},
			handlers:   successHandlers,
			wantStatus: http.StatusNotFound,
		},
		"異常ケース：パラメータにnameが無い": {
			method: http.MethodGet,
			params: map[string][]string{
				"user": {"dip 三郎"},
			},
			handlers:   successHandlers,
			wantStatus: http.StatusBadRequest,
		},
		"異常ケース：JSONエンコード失敗": {
			method: http.MethodGet,
			params: map[string][]string{
				"name": {"dip 太郎"},
			},
			resWriter:  &test.ErrorResponseWriter{},
			handlers:   successHandlers,
			wantStatus: http.StatusInternalServerError,
		},
		"異常ケース：ユーザー情報取得時にエラー発生": {
			method: http.MethodGet,
			params: map[string][]string{
				"name": {"dip 太郎"},
			},
			resWriter:  &test.ErrorResponseWriter{},
			handlers:   getUsersFailHandlers,
			wantStatus: http.StatusInternalServerError,
		},
		"異常ケース：案件情報取得時にエラー発生": {
			method: http.MethodGet,
			params: map[string][]string{
				"name": {"dip 太郎"},
			},
			resWriter:  &test.ErrorResponseWriter{},
			handlers:   getEntriesFailHandlers,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for tn, tc := range success {
		t.Run(tn, func(t *testing.T) {

			// 外部APIのモック
			ts := httptest.NewServer(test.Route(tc.handlers...))
			defer ts.Close()

			// 環境変数を一時的に変更
			oldURL := os.Getenv("MOCK_API_URL")
			os.Setenv("MOCK_API_URL", ts.URL)
			defer os.Setenv("MOCK_API_URL", oldURL)

			param := url.Values{}
			for k, p := range tc.params {
				for _, v := range p {
					param.Add(k, v)
				}
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "http://localhost/?"+param.Encode(), nil)
			Get(w, r)
			got := map[string][]Entry{}
			t.Logf(w.Body.String())
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Errorf("error: %#v, res: %#v", err, got)
			}
			assert.Equal(t, tc.wantStatus, w.Code)
			assert.Contains(t, got, keyString)
			assert.ElementsMatch(t, got[keyString], tc.response)
		})
	}
	for tn, tc := range fail {
		t.Run(tn, func(t *testing.T) {

			// 外部APIのモック
			ts := httptest.NewServer(test.Route(tc.handlers...))
			defer ts.Close()

			// 環境変数を一時的に変更
			oldURL := os.Getenv("MOCK_API_URL")
			os.Setenv("MOCK_API_URL", ts.URL)
			defer os.Setenv("MOCK_API_URL", oldURL)

			param := url.Values{}
			for k, p := range tc.params {
				for _, v := range p {
					param.Add(k, v)
				}
			}

			r := httptest.NewRequest(tc.method, "http://localhost/?"+param.Encode(), nil)

			if tc.resWriter == nil {
				w := httptest.NewRecorder()
				Get(w, r)
				assert.Equal(t, tc.wantStatus, w.Code)
			} else {
				errW := &test.ErrorResponseWriter{}
				Get(errW, r)
				assert.Equal(t, tc.wantStatus, errW.Code())
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	success := map[string]struct {
		params   map[string][]string
		response []int
	}{
		"正常ケース：データあり": {
			params: map[string][]string{
				"name": {"dip 太郎"},
			},
			response: []int{123456},
		},
		"正常ケース：データなし": {
			params: map[string][]string{
				"name": {"dip 三郎"},
			},
			response: []int{},
		},
	}
	for tn, tc := range success {
		t.Run(tn, func(t *testing.T) {

			// 外部APIのモック
			handlers := []test.Handler{
				{
					Path:    "/users",
					Handler: MockGetUser,
				},
			}
			ts := httptest.NewServer(test.Route(handlers...))
			defer ts.Close()

			// 環境変数を一時的に変更
			oldURL := os.Getenv("MOCK_API_URL")
			os.Setenv("MOCK_API_URL", ts.URL)
			defer os.Setenv("MOCK_API_URL", oldURL)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ch := make(chan []int)
			errch := make(chan error)
			go GetUserID(ctx, ch, errch, tc.params)
			select {
			case ids := <-ch:
				assert.ElementsMatch(t, ids, tc.response)
			case err := <-errch:
				t.Errorf("Error is occured : %v", err)
			}

		})
	}
}

func TestGetEntries(t *testing.T) {
	success := map[string]struct {
		params   map[string][]string
		response []Entry
	}{
		"正常ケース：データあり": {
			params: map[string][]string{
				"userID": {"123456"},
			},
			response: []Entry{
				{
					UserID: 123456,
					Name:   "案件情報1",
					Salary: 123456,
				},
			},
		},
		"正常ケース：データなし": {
			params: map[string][]string{
				"userID": {"999999"},
			},
			response: []Entry{},
		},
	}
	for tn, tc := range success {
		t.Run(tn, func(t *testing.T) {

			// 外部APIのモック
			handlers := []test.Handler{
				{
					Path:    "/entries",
					Handler: MockGetEntry,
				},
			}
			ts := httptest.NewServer(test.Route(handlers...))
			defer ts.Close()

			// 環境変数を一時的に変更
			oldURL := os.Getenv("MOCK_API_URL")
			os.Setenv("MOCK_API_URL", ts.URL)
			defer os.Setenv("MOCK_API_URL", oldURL)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ch := make(chan []Entry)
			errch := make(chan error)
			go GetEntries(ctx, ch, errch, tc.params)
			select {
			case entries := <-ch:
				assert.ElementsMatch(t, entries, tc.response)
			case err := <-errch:
				t.Errorf("Error is occured : %v", err)
			}

		})
	}
}

func MockGetUser(w http.ResponseWriter, r *http.Request) {

	// クエリ文字列に名前があるかチェック
	query := r.URL.Query()
	name := query.Get("name")

	if name == "" {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
	}

	// データを返す
	var data []User

	switch name {
	case "dip 太郎":
		data = []User{
			{
				ID:   123456,
				Name: "dip 太郎",
				Age:  25,
			},
		}
	case "dip 次郎":
		data = []User{
			{
				ID:   234567,
				Name: "dip 次郎",
				Age:  25,
			},
		}
	default:
		data = []User{}
	}

	// 値を返却する
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(data); err != nil {
		http.Error(w, "Encoding json is failed", http.StatusInternalServerError)
		return
	}

}

func MockGetEntry(w http.ResponseWriter, r *http.Request) {

	// クエリ文字列にIDがあるかチェック
	query := r.URL.Query()
	var idStrings []string
	var ok bool
	idStrings, ok = query["userID"]
	if !ok {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}
	length := len(idStrings)
	if length == 0 {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	ids := make([]int, length)
	var id int
	var err error
	for i, idString := range idStrings {
		id, err = strconv.Atoi(idString)
		if err != nil {
			http.Error(w, "Convert is failed", http.StatusInternalServerError)
			return
		}
		ids[i] = id
	}

	// データを返す
	entries := []Entry{
		{
			UserID: 123456,
			Name:   "案件情報1",
			Salary: 123456,
		},
		{
			UserID: 234567,
			Name:   "案件情報2",
			Salary: 123456,
		},
	}
	var data []Entry

	// ユーザーのIDが一致するものを探す
	for _, userID := range ids {
		for _, entry := range entries {
			if userID == entry.UserID {
				data = append(data, entry)
			}
		}
	}

	// 値を返却する
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(data); err != nil {
		http.Error(w, "Encoding json is failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}