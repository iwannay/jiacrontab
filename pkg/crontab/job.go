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
	var seconds []int
	cur := int(j.now.Second())
	if j.Second == "*" {
		cur++
		if cur > 59 {
			return 0, true, nil
		}
		return cur, false, nil
	} else if strings.Contains(j.Second, ",") {
		for _, v := range strings.Split(j.Second, ",") {
			seconds = append(seconds, util.ParseInt(v))
		}
		sort.Sort(sort.IntSlice(seconds))
		for _, v := range seconds {
			if v > cur {
				return v, false, nil
			}
		}
		return seconds[0], true, nil

	} else if strings.Contains(j.Second, "/") {
		if arr := strings.Split(j.Second, "/"); len(arr) == 2 {
			for i, j := 0, util.ParseInt(arr[1]); i <= 59; i += j {
				if i > cur {
					return i, false, nil
				}
			}
			return 0, true, nil
		}
	} else if strings.Contains(j.Second, "-") {
		if arr := strings.Split(j.Second, "-"); len(arr) == 2 {
			for i, j := util.ParseInt(arr[0]), util.ParseInt(arr[1]); i <= j; i++ {
				if i > cur {
					return i, false, nil
				}
			}
			return 0, true, nil
		}
	} else if sec, err := strconv.Atoi(j.Second); err == nil {
		if sec > cur {
			return sec, false, nil
		}
		return cur, true, nil
	}
	return 0, false, errors.New("Invalid second parameter")
}

func (j *Job) parseMinute(next bool) (int, bool, error) {
	var minutes []int
	cur := int(j.now.Minute())
	if j.Minute == "*" {

		if next {
			cur++
		}
		if cur > 59 {
			return 0, true, nil
		}

		return cur, false, nil
	} else if strings.Contains(j.Minute, ",") {
		for _, v := range strings.Split(j.Minute, ",") {
			minutes = append(minutes, util.ParseInt(v))
		}
		sort.Sort(sort.IntSlice(minutes))
		for _, v := range minutes {
			if v > cur {
				return v, false, nil
			} else if v == cur {
				if next {
					continue
				}
				return v, false, nil
			}
		}
		return minutes[0], true, nil

	} else if strings.Contains(j.Minute, "/") {
		if arr := strings.Split(j.Minute, "/"); len(arr) == 2 {

			for i, j := 0, util.ParseInt(arr[1]); i <= 59; i += j {
				if i > cur {
					return i, false, nil
				} else if i == cur {
					if next {
						continue
					}
					return i, false, nil
				}
			}
			return 0, true, nil
		}
	} else if strings.Contains(j.Minute, "-") {
		if arr := strings.Split(j.Minute, "-"); len(arr) == 2 {
			var i, j int
			for i, j = util.ParseInt(arr[0]), util.ParseInt(arr[1]); i <= j; i++ {
				if i > cur {
					return i, false, nil
				} else if i == cur {
					if next {
						continue
					}
					return i, false, nil
				}
			}
			return util.ParseInt(arr[0]), true, nil
		}
	} else if minute, err := strconv.Atoi(j.Minute); err == nil {
		if minute > cur || (minute == cur && !next) {
			return minute, false, nil
		}
		return minute, true, nil
	}

	return 0, false, errors.New("Invalid minute parameter")
}

func (j *Job) parseHour(next bool) (int, bool, error) {

	var hours []int
	cur := int(j.now.Hour())

	if j.Hour == "*" {
		if next {
			cur++
		}
		if cur > 23 {
			return 0, true, nil
		}

		return cur, false, nil
	} else if strings.Contains(j.Hour, ",") {
		for _, v := range strings.Split(j.Hour, ",") {
			hours = append(hours, util.ParseInt(v))
		}
		sort.Sort(sort.IntSlice(hours))
		for _, v := range hours {
			if v > cur {
				return v, false, nil
			} else if v == cur {
				if next {
					continue
				}
				return v, false, nil
			}
		}
		return hours[0], true, nil

	} else if strings.Contains(j.Hour, "/") {
		if arr := strings.Split(j.Hour, "/"); len(arr) == 2 {
			for i, j := 0, util.ParseInt(arr[1]); i <= 23; i += j {
				if i > cur {
					return i, false, nil
				} else if i == cur {
					if next {
						continue
					}
					return i, false, nil
				}
			}
			return 0, true, nil
		}
	} else if strings.Contains(j.Hour, "-") {
		if arr := strings.Split(j.Hour, "-"); len(arr) == 2 {
			var i, j int
			for i, j = util.ParseInt(arr[0]), util.ParseInt(arr[1]); i <= j; i++ {
				if i > cur {
					return i, false, nil
				} else if i == cur {
					if next {
						continue
					}
					return i, false, nil
				}
			}
			return util.ParseInt(arr[0]), true, nil
		}
	} else if hour, err := strconv.Atoi(j.Hour); err == nil {
		if hour > cur || (hour == cur && !next) {
			return hour, false, nil
		}
		return hour, true, nil
	}
	return 0, false, errors.New("Invalid hour parameter")
}

func (j *Job) parseDay(next bool) (int, bool, error) {

	var days []int
	cur := int(j.now.Day())
	daysNumOfMonth := util.CountDaysOfMonth(time.Now().Year(), cur)

	if j.Day == "*" {
		if next {
			cur++
		}
		if cur > daysNumOfMonth {
			return 1, true, nil
		}
		return cur, false, nil

	} else if strings.Contains(j.Day, ",") {
		for _, v := range strings.Split(j.Day, ",") {
			days = append(days, util.ParseInt(v))
		}
		sort.Sort(sort.IntSlice(days))
		for _, v := range days {
			if v > cur {
				return v, false, nil
			} else if v == cur {
				if next {
					continue
				}
				return v, false, nil
			}
		}
		return days[0], true, nil

	} else if strings.Contains(j.Day, "/") {
		if arr := strings.Split(j.Day, "/"); len(arr) == 2 {
			for i, j := 0, util.ParseInt(arr[1]); i <= daysNumOfMonth; i += j {
				if i > cur {
					return i, false, nil
				} else if i == cur {
					if next {
						continue
					}
					return i, false, nil
				}
			}
			return 0, true, nil
		}
	} else if strings.Contains(j.Day, "-") {
		if arr := strings.Split(j.Day, "-"); len(arr) == 2 {
			var i, j int
			for i, j = util.ParseInt(arr[0]), util.ParseInt(arr[1]); i <= j; i++ {
				if i > cur {
					return i, false, nil
				} else if i == cur {
					if next {
						continue
					}
					return i, false, nil
				}
			}
			return util.ParseInt(arr[0]), true, nil
		}
	} else if day, err := strconv.Atoi(j.Day); err == nil {
		if day > cur || (day == cur && !next) {
			return day, false, nil
		}

		return day, true, nil
	}
	return 0, false, errors.New("Invalid day parameter")
}

func (j *Job) parseWeekday(next bool) (int, bool, error) {
	var weekdays []int
	curWeekday := int(j.now.Weekday())
	curDay := j.now.Day()
	daysNumOfMonth := util.CountDaysOfMonth(j.now.Year(), curDay)

	if j.Weekday == "*" {
		if next {
			curDay++
		}

		if curDay > daysNumOfMonth {
			return 1, true, nil
		}

		return curDay, false, nil
	} else if strings.Contains(j.Weekday, ",") {
		for _, v := range strings.Split(j.Weekday, ",") {
			weekdays = append(weekdays, util.ParseInt(v))
		}
		sort.Sort(sort.IntSlice(weekdays))
		for _, v := range weekdays {
			if v > curWeekday {
				curDay += (v - curWeekday)
				if curDay > daysNumOfMonth {
					return weekdays[0], true, nil
				}
				return curDay, false, nil
			} else if v == curWeekday {
				if next {
					continue
				}
				return curDay, false, nil
			}
		}

		return curDay, true, nil
	} else if strings.Contains(j.Weekday, "/") {
		if arr := strings.Split(j.Weekday, "/"); len(arr) == 2 {
			for i, j := 0, util.ParseInt(arr[1]); i <= 7; i += j {
				if i > curWeekday {
					curDay += (i - curWeekday)
					if curDay > daysNumOfMonth {
						return 1, true, nil
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
				return 1, true, nil
			}
			return curDay, true, nil
		}
	} else if strings.Contains(j.Weekday, "-") {
		if arr := strings.Split(j.Weekday, "-"); len(arr) == 2 {
			var i, j int
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
	} else if weekday, err := strconv.Atoi(j.Weekday); err == nil {
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
	return 0, false, errors.New("Invalid weekday parameter")
}

func (j *Job) parseMonth(next bool) (int, bool, error) {

	var months []int
	cur := int(j.now.Month())
	daysNumOfMonth := util.CountDaysOfMonth(j.now.Year(), int(j.now.Month()))

	if j.Month == "*" {
		if next {
			cur++
		}

		if cur > daysNumOfMonth {
			return 1, true, nil
		}

		return cur, false, nil

	} else if strings.Contains(j.Second, ",") {
		for _, v := range strings.Split(j.Second, ",") {
			months = append(months, util.ParseInt(v))
		}
		sort.Sort(sort.IntSlice(months))
		for _, v := range months {
			if v > cur {
				return v, false, nil
			} else if v == cur {
				if next {
					continue
				}
				return v, false, nil
			}
		}
		return months[0], true, nil

	} else if strings.Contains(j.Month, "/") {
		if arr := strings.Split(j.Month, "/"); len(arr) == 2 {
			for i, j := 0, util.ParseInt(arr[1]); i <= 12; i += j {
				if i > cur {
					return i, false, nil
				} else if i == cur {
					if next {
						continue
					}
					return i, false, nil
				}
			}
			return 0, true, nil
		}
	} else if strings.Contains(j.Day, "-") {
		if arr := strings.Split(j.Day, "-"); len(arr) == 2 {
			var i, j int
			for i, j = util.ParseInt(arr[0]), util.ParseInt(arr[1]); i <= j; i++ {
				if i > cur {
					return i, false, nil
				} else if i == cur {
					if next {
						continue
					}
					return i, false, nil
				}
			}
			return util.ParseInt(arr[0]), true, nil
		}
	} else if month, err := strconv.Atoi(j.Month); err == nil {
		if month > cur || (month == cur && !next) {
			return month, false, nil
		}
		return cur, true, nil
	}
	return 0, false, errors.New("Invalid month parameter")
}
