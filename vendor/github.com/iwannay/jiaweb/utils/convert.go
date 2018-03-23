package utils

import (
	"encoding/json"
	"reflect"
	"strconv"
)

func Int642String(val int64) string {
	return strconv.FormatInt(val, 10)
}

func String2Int64(val string) (int64, error) {
	return strconv.ParseInt(val, 10, 64)
}

func Struct2Map(obj interface{}, out *map[string]interface{}) error {
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)

}

func Interface2Struct(in interface{}, out interface{}) error {
	var byteData []byte
	var err error
	t := reflect.TypeOf(in)

	if t.Kind() == reflect.Map {
		byteData, err = json.Marshal(in)
		if err != nil {
			return err
		}
	} else {
		byteData = in.([]byte)
	}

	return json.Unmarshal(byteData, out)
}
