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
	"time"
)

const TimeFormat = "2006-01-02T15:04:05.000000Z"

func StringToTimestamp(dateString string) (int64, error) {
	theTime, err := time.ParseInLocation(TimeFormat, dateString, time.UTC)
	if err != nil {
		return 0, err
	}
	sr := theTime.Unix()
	return sr, nil
}
