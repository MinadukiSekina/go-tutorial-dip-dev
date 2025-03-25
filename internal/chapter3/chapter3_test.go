package chapter3

import (
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

	success := map[string]struct {
		params     map[string]string
		response   []Entry
		wantStatus int
	}{
		"正常ケース": {
			params: map[string]string{
				"name": "dip 太郎",
			},
			response: []Entry{
				{
					Name:   "案件情報1",
					UserID: 123456,
					Salary: 123456,
				},
			},
			wantStatus: http.StatusOK,
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

			param := url.Values{}
			for k, v := range tc.params {
				param.Add(k, v)
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
			assert.Contains(t, keyString, got)
			assert.ElementsMatch(t, got[keyString], tc.response)
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
	idString := query.Get("id")

	if idString == "" {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	// データを返す
	var data []Entry

	switch id {
	case 123456:
		data = []Entry{
			{
				UserID: 123456,
				Name:   "案件情報1",
				Salary: 123456,
			},
		}
	case 234567:
		data = []Entry{
			{
				UserID: 234567,
				Name:   "案件情報2",
				Salary: 123456,
			},
		}
	default:
		data = []Entry{}
	}

	// 値を返却する
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(data); err != nil {
		http.Error(w, "Encoding json is failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}