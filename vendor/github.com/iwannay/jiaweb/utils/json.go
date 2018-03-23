package utils

import (
	"encoding/json"
)

func GetJsonString(obj interface{}) string {
	resByte, err := json.Marshal(obj)
	if err != nil {
		return ""
	}

	return string(resByte)
}

func Marshal(v interface{}) (string, error) {
	resByte, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(resByte), nil

}

func Unmarshal(str string, v interface{}) error {
	return json.Unmarshal([]byte(str), v)
}
