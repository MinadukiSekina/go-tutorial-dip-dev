package chapter3

import (
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
					name:   "案件情報1",
					userID: 123456,
					salary: 123456,
				},
				{
					name:   "案件情報2",
					userID: 234567,
					salary: 123456,
				},
			},
			wantStatus: http.StatusOK,
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
