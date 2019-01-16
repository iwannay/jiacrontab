package proto

const (
	SuccessRespCode = 0
	ErrorRespCode   = -1
)

type Resp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
	Sign string      `json:"sign"`
}
