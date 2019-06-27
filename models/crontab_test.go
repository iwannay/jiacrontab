package models

import (
	"encoding/json"
	"testing"
)

func TestStringSlice_Value(t *testing.T) {
	var A struct {
		Na StringSlice
		Me StringSlice
	}
	var B struct {
		Na StringSlice
		Me StringSlice
	}
	var a StringSlice
	var b StringSlice
	if a == nil {
		t.Log(len(a))
	} else {
		t.Log(2)
	}

	A.Na = make(StringSlice, 0)

	bts, err := json.Marshal(A)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(bts))

	json.Unmarshal(bts, &B)

	bts, err = json.Marshal(B)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(bts))

	a = make(StringSlice, 0)
	if a == nil {
		t.Log(3)
	} else {
		t.Log(4)
	}

	bts, err = json.Marshal(a)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(bts))
	err = json.Unmarshal([]byte("[]"), &b)
	if err != nil {
		t.Error(err)
	}
	t.Log(b, b == nil)
}
