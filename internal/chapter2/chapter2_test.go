package chapter2

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dip-dev/go-tutorial/internal/helper/test"
)

func TestMain(m *testing.M) {
	m.Run()
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
	t.Run("異常ケース:baseURLが不正", func(t *testing.T) {
		oldBaseURL := os.Getenv("MOCK_API_URL")
		defer os.Setenv("MOCK_API_URL", oldBaseURL)

		baseURL := ":\\test"
		os.Setenv("MOCK_API_URL", baseURL)

		param := url.Values{}
		param.Add("age", "25")

		r := httptest.NewRequest(http.MethodGet, "http://localhost/?"+param.Encode(), nil)
		w := httptest.NewRecorder()

		Get(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
	t.Run("異常ケース:外部APIリクエストに失敗", func(t *testing.T) {
		// エラーを返すモックサーバーを作成
		ts := httptest.NewServer(test.Route())

		defer ts.Close()

		// 環境変数を一時的に変更
		oldURL := os.Getenv("MOCK_API_URL")
		os.Setenv("MOCK_API_URL", ts.URL)
		defer os.Setenv("MOCK_API_URL", oldURL)

		param := url.Values{}
		param.Add("age", "25")

		r := httptest.NewRequest(http.MethodGet, "http://localhost/?"+param.Encode(), nil)
		w := httptest.NewRecorder()

		Get(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
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
	t.Run("異常ケース:baseURLが不正", func(t *testing.T) {
		oldBaseURL := os.Getenv("MOCK_API_URL")
		defer os.Setenv("MOCK_API_URL", oldBaseURL)

		baseURL := ":\\test"
		os.Setenv("MOCK_API_URL", baseURL)

		params, _ := json.Marshal(map[string]string{
			"name": "dip 次郎",
			"age":  "24",
		})

		r := httptest.NewRequest(http.MethodPost, "http://localhost/", bytes.NewReader(params))
		w := httptest.NewRecorder()

		Create(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
	t.Run("異常ケース:外部APIリクエストに失敗", func(t *testing.T) {
		// エラーを返すモックサーバーを作成
		ts := httptest.NewServer(test.Route())
		defer ts.Close()

		// 環境変数を一時的に変更
		oldURL := os.Getenv("MOCK_API_URL")
		os.Setenv("MOCK_API_URL", ts.URL)
		defer os.Setenv("MOCK_API_URL", oldURL)

		params, _ := json.Marshal(map[string]string{
			"name": "dip 次郎",
			"age":  "24",
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(params))

		Create(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
