package crontab

import (
	"jiacrontab/pkg/test"
	"testing"
	"time"
)

func TestJob_NextExecutionTime(t *testing.T) {
	timeLayout := "2006-01-02 15:04:05"
	j := &Job{
		Second:  "48",
		Minute:  "3",
		Hour:    "12",
		Day:     "25",
		Weekday: "*",
		Month:   "1",
	}

	tt, err := j.NextExecutionTime(time.Now())
	test.Nil(t, err)
	test.Equal(t, "2020-01-25 12:03:48", tt.Format(timeLayout))

	tt, err = j.NextExecutionTime(tt)
	test.Nil(t, err)
	test.Equal(t, "2021-01-25 12:03:48", tt.Format(timeLayout))

	tt, err = j.NextExecutionTime(tt)
	test.Equal(t, "2022-01-25 12:03:48", tt.Format(timeLayout))

	j = &Job{
		Second:  "58",
		Minute:  "*/4",
		Hour:    "12",
		Day:     "4",
		Weekday: "*",
		Month:   "3",
	}
	tt, err = j.NextExecutionTime(time.Now())
	test.Nil(t, err)
	test.Equal(t, "2020-03-04 12:00:58", tt.Format(timeLayout))

	tt, err = j.NextExecutionTime(tt)
	test.Nil(t, err)
	test.Equal(t, "2020-03-04 12:04:58", tt.Format(timeLayout))

	tt, err = j.NextExecutionTime(tt)
	test.Nil(t, err)
	test.Equal(t, "2020-03-04 12:08:58", tt.Format(timeLayout))

	j = &Job{
		Second:  "*/2",
		Minute:  "*/3",
		Hour:    "*",
		Day:     "*",
		Weekday: "*",
		Month:   "*",
	}
	tt, err = j.NextExecutionTime(time.Now())
	test.Nil(t, err)
	t.Log(tt)
	tt, err = j.NextExecutionTime(tt)
	test.Nil(t, err)
	t.Log(tt)
	tt, err = j.NextExecutionTime(tt)
	test.Nil(t, err)
	t.Log(tt)
	tt, err = j.NextExecutionTime(tt)
	test.Nil(t, err)
	t.Log(tt)
}
