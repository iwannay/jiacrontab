package proto

const (
	SuccessRespCode = 0
	ErrorRespCode   = -1
)

type Resp struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data,omitempty"`
	Sign    string      `json:"sign"`
	Version string      `json:"version"`
}
