package wxauto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
	"tools"
)

type Account struct {
	AuthInfo    AuthInfo
	BaseRequest BaseRequest
	SyncKey     SyncKey
	User        User
	MemberList  []Member
}

type SyncRequest struct {
	BaseRequest BaseRequest
	SyncKey     []Key
	rr          string
}

var escapedUserName = []string{
	"newsapp", "fmessage", "filehelper", "weibo", "qqmail", "fmessage",
	"tmessage", "qmessage", "qqsync", "floatbottle", "lbsapp", "shakeapp",
	"medianote", "qqfriend", "readerapp", "blogapp", "facebookapp",
	"masssendapp", "meishiapp", "feedsapp", "voip", "blogappweixin",
	"weixin", "brandsessionholder", "weixinreminder", "wxid_novlwrv3lqwv11",
	"gh_22b87fa7cb3c", "officialaccounts", "notification_messages",
	"wxid_novlwrv3lqwv11", "gh_22b87fa7cb3c", "wxitil", "userexperience_alarm",
	"notification_messages"}

func (this *Account) send(msg string, userName string) error {

	var sendMsgRequest SendMsgRequest
	id := fmt.Sprintf("%s", time.Now().Unix())
	sendMsgRequest.Msg = Msg{id, msg, this.User.UserName, id, userName, 1}
	sendMsgRequest.Scene = 0
	sendMsgRequest.BaseRequest = this.BaseRequest
	url := "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxsendmsg?lang=zh_CN&pass_ticket=" + this.AuthInfo.PassTicket

	body, err := tools.HttpPostJson(url, sendMsgRequest)
	if err != nil {
		return err
	}

	var sendMsgResponse SendMsgResponse
	err = json.Unmarshal(body, &sendMsgResponse)
	if err != nil || sendMsgResponse.BaseResponse.Ret != 0 {
		return err
	}

	return nil
}

func (this *Account) sendUsers(msg string, userNames []string) error {
	for _, name := range userNames {
		go this.send(msg, name)
	}
	return nil
}

func (this *Account) broadcast(msg string) error {
	for _, mem := range this.MemberList {
		if tools.InArray(mem.NickName, escapedUserName) {
			continue
		}
		go this.send(msg, mem.UserName)
	}
	return nil
}

type MsgRcv struct {
	FromUserName string
	Content      string
}

type SyncResponse struct {
	AddMsgList []MsgRcv
	SyncKey    []Key
}

func (this *Account) syn() error {

	var key_str string
	for _, v := range this.SyncKey.List {
		if len(key_str) == 0 {
			key_str += fmt.Sprintf("%d_%d", v.Key, v.Val)
		} else {
			key_str += fmt.Sprintf("|%d_%d", v.Key, v.Val)
		}

	}

	url := "https://webpush.weixin.qq.com/cgi-bin/mmwebwx-bin/synccheck?sid=" + this.AuthInfo.Wxsid
	url += "&skey=" + this.AuthInfo.Skey
	url += "&r=" + fmt.Sprintf("%s", time.Now().Unix())
	url += "&uin=" + this.AuthInfo.Wxuin
	url += "&deviceid=" + this.BaseRequest.DeviceID
	url += "&synckey=" + key_str
	url += "&_=" + fmt.Sprintf("%s", time.Now().Unix())

	body, err := tools.HttpGet(url)
	if err != nil {
		return err
	}

	////微信已退出登陆
	if bytes.Contains(body, []byte("retcode:1100,")) {
		return fmt.Errorf("微信已退出登陆")
	}

	////未知状态
	if !bytes.Contains(body, []byte("retcode:0,")) {
		return fmt.Errorf("未知状态: %s", body)
	}

	if !bytes.Contains(body, []byte("selector:2")) && !bytes.Contains(body, []byte("selector:6")) {
		return fmt.Errorf("unknown selector : %s", body)
	}

	fmt.Printf("syn_url: %s, body=%s", url, body)

	url = "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxsync?sid=" + this.AuthInfo.Wxsid
	url += "&skey=" + this.AuthInfo.Skey
	url += "&lang=zh_CN&pass_ticket=" + this.AuthInfo.PassTicket

	var syncRequest SyncRequest
	syncRequest.BaseRequest = this.BaseRequest
	syncRequest.SyncKey = this.SyncKey.List
	syncRequest.rr = fmt.Sprintf("%s", time.Now().Unix())

	body, err = tools.HttpPostJson(url, syncRequest)
	if err != nil {
		return err
	}

	fmt.Printf("syn_url: %s, body=%s", url, body)

	var syncResponse SyncResponse
	err = json.Unmarshal(body, &syncResponse)
	//if err != nil || syncResponse.BaseResponse.Ret != 0 {
	//	return err
	//}
	return nil
}
