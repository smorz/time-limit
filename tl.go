package main

import (
	"log"
	"os/exec"
	"time"

	timelimit "github.com/smorz/time-limit/pkg"
)

const (
	sessionDuration = time.Minute * 40
	check_interval  = time.Minute
	total_limit     = time.Hour * 3
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	db, err := timelimit.NewDB("data")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	baseTime, err := db.GetBaseTime()
	if err != nil {
		log.Fatal(err)
	}
	total := make(chan struct{})
	sassion := time.NewTicker(sessionDuration)
	go func() {
		select {
		//case شب بودن
		case <-total:
			log.Println("Total limit reached")
		case <-sassion.C:
			log.Println("Sassion limit reached")
		}
		Shutdown()
	}()

	check := time.NewTimer(check_interval)

	for {
		log.Println(time.Since(baseTime)) //--
		if time.Since(baseTime) >= time.Hour*24 {
			baseTime = time.Now()
			db.SetBaseTime(baseTime)
			db.SetTotalDuration(0)
		}
		td, err := db.GetTotalDuration()
		if err != nil {
			log.Fatal(err)
		}
		log.Println(td) //--
		if td >= total_limit {
			total <- struct{}{}
		}
		<-check.C
		db.IncTotalDuration(check_interval)
	}
}

func Shutdown() {
	log.Println("خاموش شد")
	if err := exec.Command("cmd", "/C", "shutdown", "/h").Run(); err != nil {
		log.Println("Failed to initiate shutdown:", err)
	}
}
