package timelimit

import (
	"fmt"
	"testing"
	"time"
)

func TestGetBaseTime(t *testing.T) {
	db, _ := NewDB("test")
	now := time.Now()
	db.SetBaseTime(now)
	ti, err := db.GetBaseTime()
	if err != nil {
		t.Error(err)
	}
	fmt.Print(ti)

	if now.Sub(ti) >= time.Second {
		t.Errorf("Difference: %v", now.Sub(ti))
	}
}

func TestGetTotalDuration(t *testing.T) {
	db, _ := NewDB("test")
	db.SetTotalDuration(time.Minute*15 + time.Second*2)
	d, err := db.GetTotalDuration()
	if err != nil {
		t.Error(err)
	}
	fmt.Print(d)

	if d != time.Minute*15+time.Second*2 {
		t.Errorf("Difference: %v", d-time.Minute*15-time.Second*2)
	}
}
