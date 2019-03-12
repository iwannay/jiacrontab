package crontab

import (
	"errors"
	"jiacrontab/pkg/util"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Job struct {
	Second            string
	Minute            string
	Hour              string
	Day               string
	Weekday           string
	Month             string
	ID                uint
	now               time.Time
	lastExecutionTime time.Time
	nextExecutionTime time.Time
	Value             interface{}
}

func (j *Job) GetNextExecTime() time.Time {
	return j.nextExecutionTime
}

func (j *Job) GetLastExecTime() time.Time {
	return j.lastExecutionTime
}

// NextExecutionTime 下次执行时间
func (j *Job) NextExecutionTime(now time.Time) (time.Time, error) {
	var (
		err                                    error
		next                                   bool
		second, minute, hour, day, month, year int
	)

	j.now = now

	second, next, err = j.parseSecond()
	if err != nil {
		return j.nextExecutionTime, err
	}

	minute, next, err = j.parseMinute(next)
	if err != nil {
		return j.nextExecutionTime, err
	}

	hour, next, err = j.parseHour(next)
	if err != nil {
		return j.nextExecutionTime, err
	}

	if j.Weekday != "*" {
		day, next, err = j.parseWeekday(next)
		if err != nil {
			return j.nextExecutionTime, err
		}
	} else {
		day, next, err = j.parseDay(next)
		if err != nil {
			return j.nextExecutionTime, err
		}
	}

	month, next, err = j.parseMonth(next)
	if err != nil {
		return j.nextExecutionTime, err
	}
	year = j.now.Year()
	if next {
		year++
	}

	j.lastExecutionTime = j.nextExecutionTime
	j.nextExecutionTime = time.Date(year, time.Month(month), day, hour, minute, second, 0, time.Local)

	return j.nextExecutionTime, nil
}

func (j *Job) parseSecond() (int, bool, error) {
	cur := int(j.now.Second())
	if j.Second == "*" {
		return validWildcardPattern(cur, 59, 0, true)
	} else if strings.Contains(j.Second, ",") {
		return validEnumPattern(j.Second, cur, true)
	} else if strings.Contains(j.Second, "/") {
		return validSplitPattern(j.Second, cur, 59, true)
	} else if strings.Contains(j.Second, "-") {
		return validSerialPattern(j.Second, cur, true)
	} else if sec, err := strconv.Atoi(j.Second); err == nil {
		return validNumPattern(cur, sec, true)
	}
	return 0, false, errors.New("Invalid second parameter")
}

func (j *Job) parseMinute(next bool) (int, bool, error) {
	cur := int(j.now.Minute())
	if j.Minute == "*" {
		return validWildcardPattern(cur, 59, 0, next)
	} else if strings.Contains(j.Minute, ",") {
		return validEnumPattern(j.Minute, cur, next)
	} else if strings.Contains(j.Minute, "/") {
		return validSplitPattern(j.Minute, cur, 59, next)
	} else if strings.Contains(j.Minute, "-") {
		return validSerialPattern(j.Minute, cur, next)
	} else if minute, err := strconv.Atoi(j.Minute); err == nil {
		return validNumPattern(cur, minute, next)
	}

	return 0, false, errors.New("Invalid minute parameter")
}

func (j *Job) parseHour(next bool) (int, bool, error) {

	cur := int(j.now.Hour())

	if j.Hour == "*" {
		return validWildcardPattern(cur, 23, 0, next)
	} else if strings.Contains(j.Hour, ",") {
		return validEnumPattern(j.Hour, cur, next)
	} else if strings.Contains(j.Hour, "/") {
		return validSplitPattern(j.Hour, cur, 23, next)
	} else if strings.Contains(j.Hour, "-") {
		return validSerialPattern(j.Hour, cur, next)
	} else if hour, err := strconv.Atoi(j.Hour); err == nil {
		return validNumPattern(cur, hour, next)
	}
	return 0, false, errors.New("Invalid hour parameter")
}

func (j *Job) parseDay(next bool) (int, bool, error) {
	cur := int(j.now.Day())
	daysNumOfMonth := util.CountDaysOfMonth(time.Now().Year(), cur)

	if j.Day == "*" {
		return validWildcardPattern(cur, daysNumOfMonth, 1, next)
	} else if strings.Contains(j.Day, ",") {
		return validEnumPattern(j.Day, cur, next)
	} else if strings.Contains(j.Day, "/") {
		return validSplitPattern(j.Day, cur, daysNumOfMonth, next)
	} else if strings.Contains(j.Day, "-") {
		return validSerialPattern(j.Day, cur, next)
	} else if j.Day == "L" {
		return validLastDay(cur, daysNumOfMonth, next)
	} else if day, err := strconv.Atoi(j.Day); err == nil {
		return validNumPattern(cur, day, next)
	}
	return 0, false, errors.New("Invalid day parameter")
}

func (j *Job) parseWeekday(next bool) (int, bool, error) {

	curWeekday := int(j.now.Weekday())
	curDay := j.now.Day()
	daysNumOfMonth := util.CountDaysOfMonth(j.now.Year(), curDay)

	if j.Weekday == "*" {
		return validWildcardPattern(curDay, daysNumOfMonth, 1, next)
	} else if strings.Contains(j.Weekday, ",") {
		return validWeekdayEnumPattern(j.Weekday, curDay, curWeekday, daysNumOfMonth, next)
	} else if strings.Contains(j.Weekday, "/") {
		return validWeekdaySplitPattern(j.Weekday, curDay, curWeekday, daysNumOfMonth, next)
	} else if strings.Contains(j.Weekday, "-") {
		return validWeekdaySerialPattern(j.Weekday, curDay, curWeekday, daysNumOfMonth, next)
	} else if weekday, err := strconv.Atoi(j.Weekday); err == nil {
		return validWeekdayNumPattern(curDay, curWeekday, weekday, daysNumOfMonth, next)
	}
	return 0, false, errors.New("Invalid weekday parameter")
}

func (j *Job) parseMonth(next bool) (int, bool, error) {
	cur := int(j.now.Month())
	if j.Month == "*" {
		return validWildcardPattern(cur, 12, 1, next)
	} else if strings.Contains(j.Second, ",") {
		return validEnumPattern(j.Second, cur, next)
	} else if strings.Contains(j.Month, "/") {
		return validSplitPattern(j.Month, cur, 12, next)
	} else if strings.Contains(j.Day, "-") {
		return validSerialPattern(j.Month, cur, next)
	} else if month, err := strconv.Atoi(j.Month); err == nil {
		return validNumPattern(cur, month, next)
	}
	return 0, false, errors.New("Invalid month parameter")
}

func parseSplitPattern(str string, defaultStart int) (start int, step int, err error) {
	arr := strings.Split(str, "/")
	if len(arr) != 2 {
		return 0, 0, errors.New("Invalid month parameter")
	}
	s := defaultStart
	e := util.ParseInt(arr[1])

	if arr[0] != "*" {
		s = util.ParseInt(arr[0])
	}
	return s, e, nil
}

func parseSerialPattern(str string) (start int, end int, err error) {
	arr := strings.Split(str, "-")
	if len(arr) != 2 {
		return 0, 0, errors.New("Invalid month parameter")
	}
	s := util.ParseInt(arr[0])
	e := util.ParseInt(arr[1])

	return s, e, nil
}

func validWildcardPattern(cur, num, start int, next bool) (int, bool, error) {
	if next {
		cur++
	}

	if cur > num {
		return start, true, nil
	}

	return cur, false, nil
}

func validSplitPattern(pattern string, cur int, num int, next bool) (int, bool, error) {

	var defaultStart int
	switch num {
	case 59, 23:
		defaultStart = 0
	default:
		defaultStart = 1
	}

	if s, e, err := parseSplitPattern(pattern, defaultStart); err == nil {
		for i, j := s, e; i <= num; i += j {
			if i > cur {
				return i, false, nil
			} else if i == cur {
				if next {
					continue
				}
				return i, false, nil
			}
		}
		return s, true, nil
	}
	return 0, false, errors.New("Invalid month parameter")
}

func validSerialPattern(pattern string, cur int, next bool) (int, bool, error) {
	if s, e, err := parseSerialPattern(pattern); err == nil {
		for i, j := s, e; i <= j; i++ {
			if i > cur {
				return i, false, nil
			} else if i == cur {
				if next {
					continue
				}
				return i, false, nil
			}
		}
		return s, true, nil
	}
	return 0, false, errors.New("Invalid month parameter")
}

func validEnumPattern(pattern string, cur int, next bool) (int, bool, error) {
	var nums []int
	for _, v := range strings.Split(pattern, ",") {
		nums = append(nums, util.ParseInt(v))
	}

	if len(nums) == 0 {
		return 0, false, errors.New("Invalid month parameter")
	}

	sort.Sort(sort.IntSlice(nums))
	for _, v := range nums {
		if v > cur {
			return v, false, nil
		} else if v == cur {
			if next {
				continue
			}
			return v, false, nil
		}
	}
	return nums[0], true, nil
}

func validWeekdayEnumPattern(pattern string, curDay, curWeekday, daysNumOfMonth int, next bool) (int, bool, error) {
	var weekdays []int
	for _, v := range strings.Split(pattern, ",") {
		weekdays = append(weekdays, util.ParseInt(v))
	}
	sort.Sort(sort.IntSlice(weekdays))
	for _, v := range weekdays {
		if v > curWeekday {
			curDay += (v - curWeekday)
			if curDay > daysNumOfMonth {
				return curDay - daysNumOfMonth, true, nil
			}
			return curDay, false, nil
		} else if v == curWeekday {
			if next {
				continue
			}
			return curDay, false, nil
		}
	}
	curDay += 7
	if curDay > daysNumOfMonth {
		return curDay - daysNumOfMonth, true, nil
	}

	return curDay, false, nil
}

func validWeekdaySplitPattern(pattern string, curDay, curWeekday, daysNumOfMonth int, next bool) (int, bool, error) {
	var weekdays []int
	if s, e, err := parseSplitPattern(pattern, 1); err == nil {
		for i, j := s, e; i <= 7; i += j {
			weekdays = append(weekdays, i)
			if i > curWeekday {
				curDay += (i - curWeekday)
				if curDay > daysNumOfMonth {
					return curDay - daysNumOfMonth, true, nil
				}
				return curDay, false, nil
			} else if i == curWeekday {
				if next {
					continue
				}
				return curDay, false, nil
			}
		}
		curDay += (7 - curWeekday + weekdays[0])
		if curDay > daysNumOfMonth {
			return curDay - daysNumOfMonth, true, nil
		}
		return curDay, false, nil
	}
	return 0, false, errors.New("Invalid weekday parameter")
}

func validWeekdaySerialPattern(pattern string, curDay, curWeekday, daysNumOfMonth int, next bool) (int, bool, error) {
	if arr := strings.Split(pattern, "-"); len(arr) == 2 {
		var i, j int
		var weekdays []int
		for i, j = util.ParseInt(arr[0]), util.ParseInt(arr[1]); i <= j; i++ {
			if i > curWeekday {
				curDay += (i - curWeekday)
				if sub := curDay - daysNumOfMonth; sub > 0 {
					return sub, true, nil
				}
				return curDay, false, nil
			} else if i == curWeekday {
				if next {
					continue
				}
				return curDay, false, nil
			}
		}
		curDay += (7 - curWeekday + weekdays[0])
		if sub := curDay - daysNumOfMonth; sub > 0 {
			return sub, true, nil
		}
		return curDay, true, nil
	}
	return 0, false, errors.New("Invalid weekday parameter")
}

func validWeekdayNumPattern(curDay, curWeekday, weekday, daysNumOfMonth int, next bool) (int, bool, error) {
	if weekday > curWeekday {
		curDay += (weekday - curWeekday)
		if sub := curDay - daysNumOfMonth; sub > 0 {
			return sub, true, nil
		}
		return curDay, false, nil
	} else if weekday == curWeekday && !next {
		return curDay, false, nil
	}

	// 跳一周
	curDay += 7
	if sub := curDay - daysNumOfMonth; sub > 0 {
		return sub, true, nil
	}
	return curDay, false, nil
}

func validNumPattern(cur, num int, next bool) (int, bool, error) {
	if num > cur || (num == cur && !next) {
		return num, false, nil
	}
	return num, true, nil
}

func validLastDay(cur, day int, next bool) (int, bool, error) {
	if day > cur || (day == cur && !next) {
		return day, false, nil
	}
	return day, true, nil
}
