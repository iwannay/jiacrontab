package handle

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"jiacrontab/model"
	"jiacrontab/pkg/log"
	"jiacrontab/pkg/proto"
	"jiacrontab/pkg/rpc"
)

func rpcCall(addr string, serviceMethod string, args interface{}, reply interface{}) error {
	err := rpc.Call(addr, serviceMethod, args, reply)
	if err != nil {
		model.DB().Unscoped().Debug().Model(&model.Client{}).Where("addr=?", addr).Update("state", 0)
	}
	return err
}

func successResp(msg string, v interface{}) proto.Resp {
	return proto.Resp{
		Code: proto.SuccessRespCode,
		Msg:  msg,
		Data: v,
	}
}

func errorResp(msg string, v interface{}) proto.Resp {

	var (
		sign string
		bts  []byte
		err  error
	)

	if msg == "" {
		msg = "success"
	}

	if v == nil {
		goto end
	}

	bts, err = json.Marshal(v)
	if err != nil {
		log.Error("errorResp:", err)
	}

	sign = fmt.Sprintf("%x", md5.Sum(bts))

end:
	return proto.Resp{
		Code: proto.ErrorRespCode,
		Msg:  msg,
		Data: string(bts),
		Sign: sign,
	}
}
