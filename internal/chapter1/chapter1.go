package chapter1

import (
	"encoding/json"
	"net/http"
	"strings"
)

func GetEcho(w http.ResponseWriter, r *http.Request) {
	//FIXME: Getメソッドのアクセスか確認
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//FIXME: パラメータをFormに変換する
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//FIXME: パラメータを取得する
	var ps = map[string]string{}
	for k, v := range r.Form {
		ps[k] = strings.Join(v, "")
	}

	//FIXME: パラメータをレスポンスに書き出す
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(ps); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//FIXME: レスポンスコード設定
	w.WriteHeader(http.StatusOK)
}
