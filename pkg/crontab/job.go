// package crontab 实现定时调度
// 借鉴https://github.com/robfig/cron
// 部分实现添加注释
// 向https://github.com/robfig/cron项目致敬
package crontab

import (
	"errors"
	"jiacrontab/pkg/util"
	"time"
)

const (
	starBit = 1 << 63
)

type bounds struct {
	min, max uint
	names    map[string]uint
}

// The bounds for each field.
var (
	seconds = bounds{0, 59, nil}
	minutes = bounds{0, 59, nil}
	hours   = bounds{0, 23, nil}
	dom     = bounds{1, 31, nil}
	months  = bounds{1, 12, map[string]uint{
		"jan": 1,
		"feb": 2,
		"mar": 3,
		"apr": 4,
		"may": 5,
		"jun": 6,
		"jul": 7,
		"aug": 8,
		"sep": 9,
		"oct": 10,
		"nov": 11,
		"dec": 12,
	}}
	dow = bounds{0, 6, map[string]uint{
		"sun": 0,
		"mon": 1,
		"tue": 2,
		"wed": 3,
		"thu": 4,
		"fri": 5,
		"sat": 6,
	}}
)

type Job struct {
	Second  string
	Minute  string
	Hour    string
	Day     string
	Weekday string
	Month   string

	ID                uint
	now               time.Time
	lastExecutionTime time.Time
	nextExecutionTime time.Time

	second, minute, hour, dom, month, dow uint64

	Value interface{}
}

func (j *Job) GetNextExecTime() time.Time {
	return j.nextExecutionTime
}

func (j *Job) GetLastExecTime() time.Time {
	return j.lastExecutionTime
}

// parse 解析定时规则
// 根据规则生成符和条件的日期
// 例如：*/2 如果位于分位，则生成0,2,4,6....58
// 生成的日期逐条的被映射到uint64数值中
// min |= 1<<2
func (j *Job) parse() error {
	var err error
	field := func(field string, r bounds) uint64 {
		if err != nil {
			return 0
		}
		var bits uint64
		bits, err = getField(field, r)
		return bits
	}
	j.second = field(j.Second, seconds)
	j.minute = field(j.Minute, minutes)
	j.hour = field(j.Hour, hours)
	j.dom = field(j.Day, dom)
	j.month = field(j.Month, months)
	j.dow = field(j.Weekday, dow)

	return err

}

// NextExecTime 获得下次执行时间
func (j *Job) NextExecutionTime(t time.Time) (time.Time, error) {
	if err := j.parse(); err != nil {
		return time.Time{}, err
	}

	t = t.Add(1*time.Second - time.Duration(t.Nanosecond())*time.Nanosecond)
	added := false
	defer func() {
		j.lastExecutionTime, j.nextExecutionTime = j.nextExecutionTime, t
	}()

	// 设置最大调度周期为5年
	yearLimit := t.Year() + 5

WRAP:
	if t.Year() > yearLimit {
		return time.Time{}, errors.New("Over 5 years")
	}

	for 1<<uint(t.Month())&j.month == 0 {
		// If we have to add a month, reset the other parts to 0.
		if !added {
			added = true
			// Otherwise, set the date at the beginning (since the current time is irrelevant).
			t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		}
		t = t.AddDate(0, 1, 0)

		// Wrapped around.
		if t.Month() == time.January {
			goto WRAP
		}
	}

	// Now get a day in that month.
	for !dayMatches(j, t) {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		}
		t = t.AddDate(0, 0, 1)

		if t.Day() == 1 {
			goto WRAP
		}
	}

	for 1<<uint(t.Hour())&j.hour == 0 {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
		}
		t = t.Add(1 * time.Hour)

		if t.Hour() == 0 {
			goto WRAP
		}
	}

	for 1<<uint(t.Minute())&j.minute == 0 {
		if !added {
			added = true
			t = t.Truncate(time.Minute)
		}
		t = t.Add(1 * time.Minute)

		if t.Minute() == 0 {
			goto WRAP
		}
	}

	for 1<<uint(t.Second())&j.second == 0 {
		if !added {
			added = true
			t = t.Truncate(time.Second)
		}
		t = t.Add(1 * time.Second)

		if t.Second() == 0 {
			goto WRAP
		}
	}

	return t, nil
}

func dayMatches(j *Job, t time.Time) bool {

	if j.Day == "L" {
		l := util.CountDaysOfMonth(t.Year(), int(t.Month()))
		j.dom = getBits(uint(l), uint(l), 1)
	}

	var (
		domMatch bool = 1<<uint(t.Day())&j.dom > 0
		dowMatch bool = 1<<uint(t.Weekday())&j.dow > 0
	)

	if j.dom&starBit > 0 || j.dow&starBit > 0 {
		return domMatch && dowMatch
	}
	return domMatch || dowMatch
}
