package main

import (
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/nathan-osman/go-sunrise"
	"github.com/smorz/time-limit/database"
)

const (
	checkInterval                    = time.Minute
	allowedTimeInOneCycle            = time.Minute * 150
	oneCycle                         = 14*time.Hour + 35*time.Minute
	allowedTimeForOneSession         = time.Minute * 50
	necessaryRestUntilTheNextSession = time.Minute * 30
	Latitude                         = 35.6892
	Longitude                        = 51.3890

	cycleStartKey                  = "cycle_start"
	sinceTheBeginningOfTheCycleKey = "since_the_beginning_of_the_cycle"
	lastTimeOnKey                  = "last_time_on"
	sinceTheStartOfTheSessionKey   = "since_the_start_of_the_session"
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	logFile, err := os.OpenFile("time-limit.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)
	log.Println("Start")
	db, err := database.NewDB("data")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	cycleStart, err := db.GetTime(cycleStartKey)
	if err != nil {
		log.Fatal(err)
	}

	check := time.NewTicker(checkInterval)
	var restarted bool

	for {
		if IsNight() {
			log.Println("ّIt is night!")
			Shutdown()
		}

		sinceTheStartOfTheSession, err := db.GetDuration(sinceTheStartOfTheSessionKey)
		if err != nil {
			log.Fatal(err)
		}

		// Has a new session started?
		lastTimeOn, err := db.GetTime(lastTimeOnKey)
		if err != nil {
			log.Fatal(err)
		}
		if sinceTheLastTimeOn := time.Since(lastTimeOn); sinceTheLastTimeOn > checkInterval*2 {
			restarted = true
			log.Printf("Was stopped at %v.\n", lastTimeOn)
			log.Println("Restart")
			if sinceTheLastTimeOn > necessaryRestUntilTheNextSession {
				// Reset the usage rate since the start of a session.
				sinceTheStartOfTheSession = 0
			} else {
				if sinceTheLastTimeOn > necessaryRestUntilTheNextSession/2 {
					// It is a good idea to reduce the off time from the session time.
					sinceTheStartOfTheSession -= sinceTheLastTimeOn * allowedTimeForOneSession / necessaryRestUntilTheNextSession
					if sinceTheStartOfTheSession < 0 {
						sinceTheStartOfTheSession = 0
					}
				}
			}
			db.SetDuration(sinceTheStartOfTheSessionKey, sinceTheStartOfTheSession)
		}
		// Is a cycle over?
		if time.Since(cycleStart) >= oneCycle {
			cycleStart = time.Now()
			db.SetTime(cycleStartKey, cycleStart)
			// Reset the usage rate since the start of a cycle.
			db.SetDuration(sinceTheBeginningOfTheCycleKey, 0)
		}

		// Has the usage rate reached the maximum allowed since the beginning of the session?
		if sinceTheStartOfTheSession >= allowedTimeForOneSession {
			log.Println("Reached the maximum time allowed for one session.")
			Shutdown()
		}

		//Has the usage rate reached the maximum allowed since the beginning of the cycle?
		sinceTheBeginningOfTheCycle, err := db.GetDuration(sinceTheBeginningOfTheCycleKey)
		if err != nil {
			log.Fatal(err)
		}
		if sinceTheBeginningOfTheCycle >= allowedTimeInOneCycle {
			log.Println("Reached the maximum time allowed for one cycle.")
			Shutdown()
		}

		<-check.C
		// Increase durations
		db.IncDuration(sinceTheBeginningOfTheCycleKey, checkInterval)
		db.IncDuration(sinceTheStartOfTheSessionKey, checkInterval)

		// Update the last time it was on
		if restarted || time.Since(lastTimeOn) < checkInterval*2 {
			db.SetTime(lastTimeOnKey, time.Now())
			restarted = false
		}

	}
}

func Shutdown() {
	log.Println("Shutdown")
	if err := exec.Command("cmd", "/C", "shutdown", "/s").Run(); err != nil {
		log.Println("Failed to initiate shutdown:", err)
	}
}

func IsNight() bool {
	now := time.Now()
	r, s := sunrise.SunriseSunset(Latitude, Longitude, now.Year(), now.Month(), now.Day())
	return !(now.After(r.Local()) && now.Before(s.Local()))
}
