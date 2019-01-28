package crontab

import (
	"jiacrontab/pkg/util"
	"sort"
	"strings"
	"time"
)

type Job struct {
	Second            string
	Minute            string
	Hour              string
	Day               string
	Week              string
	Month             string
	ID                int
	lastExecutionTime time.Time
	nextExecutionTime time.Time
	Value             interface{}
}

// NextExecutionTime 下次执行时间
func (j *Job) NextExecutionTime() (time.Time, bool) {
	var t time.Time
	second, ok := j.parseSecond()
	if !ok {
		return t, ok
	}
	minute, ok := j.parseMinute()
	if !ok {
		return t, ok
	}

	hour, ok := j.parseHour()
	if !ok {
		return t, ok
	}

	day, ok := j.parseDay()
	if !ok {
		return t, ok
	}

	_, ok := j.parseWeekday()
	if !ok {
		return t, ok
	}

	month, ok := j.parseMonth()
	if !ok {
		return t, ok
	}

	time.Date(time.Now().Year(), time.Month(month), day, hour, minute, second, 0, time.Local)

	return t, ok

}

func (j *Job) parseSecond() (int, bool) {
	var seconds []int
	cur := int(time.Now().Second())

	if j.Second == "*" {
		return cur + 1, true
	} else if strings.Contains(j.Second, ",") {

		for _, v := range strings.Split(j.Second, ",") {
			seconds = append(seconds, util.ParseInt(v))
		}
		sort.Sort(sort.IntSlice(seconds))
		for _, v := range seconds {
			if v == cur {
				return cur, true
			} else if v >= cur {
				return v, true
			}
		}

	} else if strings.Contains(j.Second, "/") {
		if arr := strings.Split(j.Second, "/"); len(arr) == 2 {
			return time.Now().Add(time.Second * time.Duration(util.ParseInt(arr[1]))).Second(), true
		}
	} else if strings.Contains(j.Second, "-") {
		if arr := strings.Split(j.Second, "/"); len(arr) == 2 {
			for i, j := util.ParseInt(arr[0]), util.ParseInt(arr[1]); i <= j; i++ {
				if cur < i {
					return i, true
				}
			}
		}
	}
	return 0, false
}

func (j *Job) parseMinute() (int, bool) {
	return 0, true
}

func (j *Job) parseHour() (int, bool) {
	return 0, true
}

func (j *Job) parseDay() (int, bool) {
	return 0, true
}

func (j *Job) parseWeekday() (int, bool) {
	return 0, true
}

func (j *Job) parseMonth() (int, bool) {
	return 0, true
}
