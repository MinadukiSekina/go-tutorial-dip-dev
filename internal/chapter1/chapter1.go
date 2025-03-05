package chapter1

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func GetEcho(w http.ResponseWriter, r *http.Request) {
	//FIXME: Getメソッドのアクセスか確認
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//FIXME: パラメータをFormに変換する
	if err := r.ParseForm(); err != nil {
		fmt.Println("Parse error!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//FIXME: パラメータを取得する
	var ps = map[string]string{}
	for k, v := range r.Form {
		ps[k] = v[0]
	}

	//FIXME: パラメータをレスポンスに書き出す
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(ps); err != nil {
		return
	}

	//FIXME: レスポンスコード設定
	w.WriteHeader(http.StatusOK)
}
