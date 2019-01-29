package crontab

import (
	"reflect"
	"testing"
	"time"
)

func TestJob_NextExecutionTime(t *testing.T) {
	type fields struct {
		Second            string
		Minute            string
		Hour              string
		Day               string
		Weekday           string
		Month             string
		ID                int
		lastExecutionTime time.Time
		nextExecutionTime time.Time
		Value             interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		want    time.Time
		wantErr bool
	}{
		{
			name: "1",
			fields: fields{
				Second:  "48",
				Minute:  "3",
				Hour:    "12",
				Day:     "25",
				Weekday: "*",
				Month:   "1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &Job{
				Second:            tt.fields.Second,
				Minute:            tt.fields.Minute,
				Hour:              tt.fields.Hour,
				Day:               tt.fields.Day,
				Weekday:           tt.fields.Weekday,
				Month:             tt.fields.Month,
				ID:                tt.fields.ID,
				lastExecutionTime: tt.fields.lastExecutionTime,
				nextExecutionTime: tt.fields.nextExecutionTime,
				Value:             tt.fields.Value,
			}
			got, err := j.NextExecutionTime()
			if (err != nil) != tt.wantErr {
				t.Errorf("Job.NextExecutionTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Job.NextExecutionTime() = %v, want %v", got.Format("2006-01-02 15:04:05"), tt.want)
			}
		})
	}
}
