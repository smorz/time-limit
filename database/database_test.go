package database

import (
	"fmt"
	"testing"
	"time"
)

func TestGetTime(t *testing.T) {
	db, _ := OpenDB("test")
	now := time.Now()
	db.Set("base_time", now)
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
	db, _ := OpenDB("test")
	db.Set("total_duration", time.Minute*15+time.Second*2)
	d, err := db.GetDuration("total_duration")
	if err != nil {
		t.Error(err)
	}
	fmt.Print(d)

	if d != time.Minute*15+time.Second*2 {
		t.Errorf("Difference: %v", d-time.Minute*15-time.Second*2)
	}
}
