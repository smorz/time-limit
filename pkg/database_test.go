package database

import (
	"fmt"
	"testing"
	"time"
)

func TestGetTime(t *testing.T) {
	db, _ := NewDB("test")
	now := time.Now()
	db.SetTime("base_time", now)
	ti, err := db.GetTime("base_time")
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
	db.SetDuration("total_duration", time.Minute*15+time.Second*2)
	d, err := db.GetDuration("total_duration")
	if err != nil {
		t.Error(err)
	}
	fmt.Print(d)

	if d != time.Minute*15+time.Second*2 {
		t.Errorf("Difference: %v", d-time.Minute*15-time.Second*2)
	}
}
