package log

import "testing"

func TestInfo(t *testing.T) {
	SetOptput(NewWriter(&WriterOptions{
		Dir:    "logfiles",
		Size:   100,
		Prefix: "2006-01-02.",
		Suffix: ".log",
	}))

	for i := 0; i < 100; i++ {
		Info("hello boy")
		Error("hello boy")
	}

}

func TestInfo2(t *testing.T) {
	for i := 0; i < 100; i++ {
		Info("hello boy")
		Error("hello boy")
	}

}
