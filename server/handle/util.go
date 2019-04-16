package handle

import (
	"database/sql"
	"jiacrontab/libs/rpc"
	"jiacrontab/model"
)

func rpcCall(addr string, serviceMethod string, args interface{}, reply interface{}) error {
	err := rpc.Call(addr, serviceMethod, args, reply)
	if err != nil && err != sql.ErrNoRows {
		model.DB().Unscoped().Debug().Model(&model.Client{}).Where("addr=?", addr).Update("state", 0)
	}
	return err
}
