// WeChatAuto project main.go
package main

import (
	"log"
	"net/http"
	"wxauto"
)

var wxClient wxauto.WxClient

func gInit() {
	if wxClient.OnlineAccount == nil {
		wxClient.OnlineAccount = make(map[string]wxauto.Account)
	}
}

func main() {
	gInit()

	http.HandleFunc("/", wxClient.Index)
	http.HandleFunc("/loginCode", wxClient.GetLoginCode)
	http.HandleFunc("/login", wxClient.LoginScan)
	http.HandleFunc("/sendMessage", wxClient.SendMsg)
	http.HandleFunc("/broadcastMessage", wxClient.BroadcastMsg)
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}
