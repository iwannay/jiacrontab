package proto

const (
	Code_Success    = 0
	Code_FailedAuth = 5001
	Code_Error      = 5002
	Code_NotFound   = 5004
	Code_NotAllowed = 5005
	Code_JWTError   = 5006
)

const (
	Msg_NotAllowed = "permission not allowed"
	Msg_JWTError   = "parse jwt failed"
)
