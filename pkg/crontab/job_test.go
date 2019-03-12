package crontab

import (
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
	t.Logf("job1 1 next:%s, %v", tt.Format(timeLayout), err)

	tt, err = j.NextExecutionTime(tt)
	t.Logf("job1 2 next:%s, %v", tt.Format(timeLayout), err)

	tt, err = j.NextExecutionTime(tt)
	t.Logf("job1 3 next:%s, %v", tt.Format(timeLayout), err)

	j = &Job{
		Second:  "58",
		Minute:  "*/4",
		Hour:    "12",
		Day:     "4",
		Weekday: "*",
		Month:   "3",
	}
	tt, err = j.NextExecutionTime(time.Now())
	t.Logf("job2 1 next:%s, %v", tt.Format(timeLayout), err)

	tt, err = j.NextExecutionTime(tt)
	t.Logf("job2 2 next:%s, %v", tt.Format(timeLayout), err)

	tt, err = j.NextExecutionTime(tt)
	t.Logf("job2 3 next:%s, %v", tt.Format(timeLayout), err)

	j = &Job{
		Second:  "50",
		Minute:  "6",
		Hour:    "*",
		Day:     "L",
		Weekday: "*",
		Month:   "*",
	}
	tt, err = j.NextExecutionTime(time.Now())
	t.Logf("job3 1 next:%s, %v", tt.Format(timeLayout), err)

}
