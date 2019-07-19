/*
 * Copyright (C) 2017. Huawei Technologies Co., LTD. All rights reserved.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of Apache License, Version 2.0.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * Apache License, Version 2.0 for more details.
 */
package util

import (
	"encoding/json"
	"net/url"
	"io"
	"strings"
	"fmt"
)

// get the request query params
func GetQueryParams(request interface{}) (urlEncoded string, err error) {
	jsonStr, err := json.Marshal(request)
	if err != nil {
		return
	}
	var result map[string]interface{}
	if err = json.Unmarshal(jsonStr, &result); err != nil {
		return
	}

	urlEncoder := url.Values{}
	for key, value := range result {
		str := fmt.Sprint(value)
		if str != "" {
			urlEncoder.Add(key, str)
		}
	}
	urlEncoded = urlEncoder.Encode()
	return
}

// get the request body params
func GetBodyParams(request interface{}) (reader io.Reader, err error) {
	jsonStr, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	return strings.NewReader(string(jsonStr)), nil
}
