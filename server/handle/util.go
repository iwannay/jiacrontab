package handle

import (
	"jiacrontab/libs/rpc"
	"jiacrontab/model"
)

func rpcCall(addr string, serviceMethod string, args interface{}, reply interface{}) error {
	err := rpc.Call(addr, serviceMethod, args, reply)
	if err != nil {
		model.DB().Model(&model.Client{}).Where("addr", addr).Update("state", 0)
	}
	return err
}
