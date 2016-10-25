package wxauto

type DataReturn struct {
	Code string
	Msg  string
	Data interface{}
}

type BaseRequest struct {
	DeviceID string
	Sid      string
	Skey     string
	Uin      string
}

type BaseResponse struct {
	Ret    int
	ErrMsg string
}

type QrImageResult struct {
	LoginCode string
	ImgUrl    string
}
type AuthInfo struct {
	Skey       string
	Wxsid      string
	Wxuin      string
	PassTicket string
}

type StatusNotifyRequest struct {
	BaseRequest  BaseRequest
	Code         int
	FromUserName string
	ToUserName   string
	ClientMsgId  string
}
type UserInfoRequest struct {
	BaseRequest BaseRequest
}

type Key struct {
	Key int
	Val int64
}
type SyncKey struct {
	Count int
	List  []Key
}

type User struct {
	Uin        int64
	UserName   string
	NickName   string
	HeadImgUrl string
}

type UserRespond struct {
	BaseResponse BaseResponse
	SyncKey      SyncKey
	User         User
}

type Member struct {
	VerifyFlag int
	UserName   string
	RemarkName string
	NickName   string
}

type ContactsRespond struct {
	BaseResponse BaseResponse
	MemberList   []Member
}

type UserAndContacts struct {
	User       *User
	MemberList []Member
}

type Msg struct {
	ClientMsgId  string
	Content      string
	FromUserName string
	LocalID      string
	ToUserName   string
	Type         int
}

type SendMsgRequest struct {
	Msg         Msg
	Scene       int
	BaseRequest BaseRequest
}

type SendMsgResponse struct {
	BaseResponse BaseResponse
}
