package chapter2

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
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
}
