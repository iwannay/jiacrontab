package admin

import (
	"jiacrontab/model"
	"jiacrontab/models"
	"jiacrontab/pkg/rpc"
)

func rpcCall(addr string, serviceMethod string, args interface{}, reply interface{}) error {
	err := rpc.Call(addr, serviceMethod, args, reply)
	if err != nil {
		models.DB().Unscoped().Debug().Model(&model.Client{}).Where("addr=?", addr).Update("state", 0)
	}
	return err
}
