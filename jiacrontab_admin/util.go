package admin

import (
	"jiacrontab/models"
	"github.com/iwannay/log"
	"jiacrontab/pkg/rpc"
)

func rpcCall(addr string, serviceMethod string, args interface{}, reply interface{}) error {
	err := rpc.Call(addr, serviceMethod, args, reply)
	if err != nil {
		ret := models.DB().Unscoped().Model(&models.Node{}).Where("addr=?", addr).Update("disabled", true)
		if ret.Error != nil {
			log.Errorf("rpcCall:%v", ret.Error)
		}
	}
	return err
}
