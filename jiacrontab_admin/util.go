package admin

import (
	"errors"
	"jiacrontab/models"
	"jiacrontab/pkg/rpc"
	"reflect"
	"strings"

	"github.com/iwannay/log"
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

func validStructRule(i interface{}) error {
	rt := reflect.TypeOf(i)
	rv := reflect.ValueOf(i)

	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		r := sf.Tag.Get("rule")
		br := sf.Tag.Get("bind")

		if br == "required" && rv.Field(i).Kind() == reflect.Ptr {
			if rv.Field(i).IsNil() {
				return errors.New(sf.Name + " is required")
			}
		}

		if r == "" {
			continue
		}
		if rs := strings.Split(r, ","); len(rs) == 2 {
			rvf := rv.Field(i)
			if rs[0] == "required" {
				switch rvf.Kind() {
				case reflect.String:
					if rvf.Interface() == "" {
						return errors.New(rs[1])
					}
				case reflect.Array, reflect.Map:
					if rvf.Len() == 0 {
						return errors.New(rs[1])
					}

				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if rvf.Interface() == 0 {
						return errors.New(rs[1])
					}
				default:
				}

			}
			continue
		}
	}
	return nil
}
