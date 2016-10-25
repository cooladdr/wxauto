package wxauto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"tools"
)

type WxClient struct {
	OnlineAccount map[string]Account
}

func qrScan(uuid string) (redirect string, err error) {

	url := "https://login.weixin.qq.com/cgi-bin/mmwebwx-bin/login?uuid=" + uuid
	url += "&tip=1&_=" + strconv.FormatInt(time.Now().Unix(), 10)

	body, err := tools.HttpGet(url)
	if err != nil {
		return "", err
	}

	if bytes.Contains(body, []byte("window.code=201")) {
		//window.code=201;----扫描了二维码但没有确认登陆
		return "scaned", nil

	} else if bytes.Contains(body, []byte("window.code=200")) {
		//window.code=200;----确认登陆成功
		//window.redirect_uri="https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxnewloginpage?
		//ticket=AUh8EKTqeDWg8Wgt_8ncCAn-@qrticket_0&uuid=IfjAbpvrQQ==&lang=zh_CN&scan=1469532226";

		idx := bytes.Index(body, []byte("window.redirect_uri="))
		start := idx + len("window.redirect_uri=") + 1
		end := len(body) - 2
		redirectUrl := body[start:end]

		return string(redirectUrl) + "&fun=new&version=v2", nil

	} else {
		//----未知返回数据 如： window.code=408;
		return "", fmt.Errorf("wx unknow return: " + string(body))
	}
}

func wxInit(url string) (auth *AuthInfo, err error) {

	body, err := tools.HttpGet(url)
	if err != nil {
		return nil, err
	}

	//<error><ret>0</ret><message>OK</message>
	//<skey>@crypt_82c272e_a53b6ff74ee689c094ec83facfa53c12</skey>
	//<wxsid>y9mDW5qqRmhy6Fmo</wxsid>
	//<wxuin>2964982861</wxuin>
	//<pass_ticket>L2rsfv51e2Uh7G9orhDcAvVBNXEAoxtVQg3cBzyJ%2BjzHZZaACOg2gL7uNNgd2q%2FJ</pass_ticket>
	//<isgrayscale>1</isgrayscale></error>
	if bytes.Contains(body, []byte("<ret>0</ret>")) {
		auth = new(AuthInfo)
		idxStart := bytes.Index(body, []byte("<skey>"))
		idxEnd := bytes.Index(body, []byte("</skey>"))
		auth.Skey = string(body[idxStart+len("<skey>") : idxEnd])

		subBody := body[idxEnd:]
		idxStart = bytes.Index(subBody, []byte("<wxsid>"))
		idxEnd = bytes.Index(subBody, []byte("</wxsid>"))
		auth.Wxsid = string(subBody[idxStart+len("<wxsid>") : idxEnd])

		subBody = subBody[idxEnd:]
		idxStart = bytes.Index(subBody, []byte("<wxuin>"))
		idxEnd = bytes.Index(subBody, []byte("</wxuin>"))
		auth.Wxuin = string(subBody[idxStart+len("<wxuin>") : idxEnd])

		subBody = subBody[idxEnd:]
		idxStart = bytes.Index(subBody, []byte("<pass_ticket>"))
		idxEnd = bytes.Index(subBody, []byte("</pass_ticket>"))
		auth.PassTicket = string(subBody[idxStart+len("<pass_ticket>") : idxEnd])

		return auth, nil
	}

	return nil, fmt.Errorf("wxInit incorrect return : " + string(body))
}

func statusNotify(baseRequest *BaseRequest, authInfo *AuthInfo, userName string) error {
	url := "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxstatusnotify?lang=zh_CN&pass_ticket=" + authInfo.PassTicket

	statusNotifyRequest := StatusNotifyRequest{
		*baseRequest,
		3,
		userName,
		userName,
		fmt.Sprintf("%s", time.Now().Unix())}

	_, err := tools.HttpPostJson(url, statusNotifyRequest)

	return err
}

func getContacts(baseRequest *BaseRequest, authInfo *AuthInfo) ([]byte, error) {
	url := "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxgetcontact?lang=zh_CN&"
	url += "r=" + fmt.Sprintf("%s", time.Now().Unix())
	url += "&lang=zh_CN&pass_ticket=" + authInfo.PassTicket
	url += "&seq=0&skey=" + authInfo.Skey

	userInfoRequest := UserInfoRequest{*baseRequest}
	body, err := tools.HttpPostJson(url, userInfoRequest)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func getUserInfo(baseRequest *BaseRequest, authInfo *AuthInfo) ([]byte, error) {
	url := "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxinit?"
	url += "r=" + fmt.Sprintf("%s", time.Now().Unix())
	url += "&lang=zh_CN&pass_ticket=" + authInfo.PassTicket

	userInfoRequest := UserInfoRequest{*baseRequest}
	body, err := tools.HttpPostJson(url, userInfoRequest)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (this *WxClient) GetLoginCode(w http.ResponseWriter, r *http.Request) {
	retData := DataReturn{Code: "-1"}

	url := "https://login.weixin.qq.com/jslogin?"
	url += "appid=wx782c26e4c19acffb&"
	url += "redirect_uri=https%3A%2F%2Fwx.qq.com%2Fcgi-bin%2Fmmwebwx-bin%2Fwebwxnewloginpage&"
	url += "fun=new&"
	url += "lang=zh_CN&_=" + fmt.Sprintf("%s", time.Now().Unix())

	body, err := tools.HttpGet(url)
	if err != nil {
		fmt.Printf("%v", err)
		tools.JsonReturn(w, retData)
		return
	}

	//body : window.QRLogin.code = 200; window.QRLogin.uuid = "wfggmPdqLA==";
	if bytes.Contains(body, []byte("code = 200")) {
		idx := bytes.Index(body, []byte("uuid = "))
		start := idx + len("uuid = ") + 1
		end := len(body) - 2
		uuid := body[start:end]

		var qrResult QrImageResult
		qrResult.LoginCode = string(uuid)
		qrResult.ImgUrl = "http://login.weixin.qq.com/qrcode/" + qrResult.LoginCode + "?t=webwx"

		retData.Code = "0"
		retData.Msg = "success"
		retData.Data = qrResult
	}

	tools.JsonReturn(w, retData)
	return
}

func (this *WxClient) LoginScan(w http.ResponseWriter, r *http.Request) {
	var redirect string
	var err error
	retData := DataReturn{Code: "-1"}

	uuid := r.URL.Query().Get("uuid")
	if !(len(uuid) > 0) {
		retData.Msg = "No uuid provided"
		tools.JsonReturn(w, retData)
		return
	}

	var deviceId = "e" + fmt.Sprintf("%s", time.Now().Unix())

	for {
		redirect, err = qrScan(uuid)
		if err != nil {
			fmt.Printf("%v", err)
			break
		}

		//用户已经扫描了二维码，但未确认登陆
		if redirect == "scaned" {
			time.Sleep(time.Second)
			continue
		}

		///初始化登陆，并得到一些授权信息
		authInfo, err := wxInit(redirect)
		if err != nil {
			retData.Msg = "reload"
			break
		}

		//组装请求参数
		var baseRequest BaseRequest
		baseRequest.DeviceID = deviceId
		baseRequest.Sid = authInfo.Wxsid
		baseRequest.Skey = authInfo.Skey
		baseRequest.Uin = authInfo.Wxuin

		//读取用户的基本信息
		userData, err := getUserInfo(&baseRequest, authInfo)
		if err != nil {
			fmt.Printf("%s", err)
			retData.Msg = "reload"
			break
		}

		var userInfo UserRespond
		err = json.Unmarshal(userData, &userInfo)
		if err != nil || userInfo.BaseResponse.Ret != 0 {
			fmt.Printf("%s", err)
			retData.Msg = "reload"
			break
		}

		//向微信服务器确认已经收到 用户信息
		_ = statusNotify(&baseRequest, authInfo, userInfo.User.UserName)

		//读取用户的好友列表
		contactsData, err := getContacts(&baseRequest, authInfo)
		if err != nil {
			fmt.Printf("%s", err)
			retData.Msg = "reload"
			break
		}

		var contactsRespond ContactsRespond
		err = json.Unmarshal(contactsData, &contactsRespond)
		if err != nil || contactsRespond.BaseResponse.Ret != 0 {
			fmt.Printf("%s", err)
			retData.Msg = "reload"
			break
		}

		var userAndContacts UserAndContacts
		userAndContacts.User = &userInfo.User
		userAndContacts.MemberList = contactsRespond.MemberList

		retData.Code = "0"
		retData.Data = userAndContacts

		//将用户加入登陆列表
		this.OnlineAccount[userInfo.User.NickName] = Account{
			*authInfo,
			baseRequest,
			userInfo.SyncKey,
			userInfo.User,
			contactsRespond.MemberList}

		account := this.OnlineAccount[userInfo.User.NickName]
		err = account.syn()
		fmt.Printf("%s", err)
		break
	}

	tools.JsonReturn(w, retData)
	return
}

func (this *WxClient) Index(w http.ResponseWriter, r *http.Request) {

}

func (this *WxClient) SendMsg(w http.ResponseWriter, r *http.Request) {

}

func (this *WxClient) BroadcastMsg(w http.ResponseWriter, r *http.Request) {

}
