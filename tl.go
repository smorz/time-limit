package main

import (
	"log"
	"os/exec"
	"time"

	database "github.com/smorz/time-limit/pkg"
)

const (
	checkInterval                    = time.Minute
	allowedTimeInOneCycle            = time.Minute * 150
	oneCycle                         = time.Hour * 24
	allowedTimeForOneSession         = time.Minute * 50
	necessaryRestUntilTheNextSession = time.Minute * 30

	cycleStartKey                  = "cycle_start"
	sinceTheBeginningOfTheCycleKey = "since_the_beginning_of_the_cycle"
	lastTimeOnKey                  = "last_time_on"
	sinceTheStartOfTheSessionKey   = "since_the_start_of_the_session"
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	db, err := database.NewDB("data")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	cycleStart, err := db.GetTime(cycleStartKey)
	if err != nil {
		log.Fatal(err)
	}
	sinceTheStartOfTheSession, err := db.GetDuration(sinceTheStartOfTheSessionKey)
	if err != nil {
		log.Fatal(err)
	}
	lastTimeOn, err := db.GetTime(lastTimeOnKey)
	if err != nil {
		log.Fatal(err)
	}

	// It is a good idea to reduce the off time from the session time.
	sinceTheLastTimeOn := time.Since(lastTimeOn)
	if sinceTheLastTimeOn > necessaryRestUntilTheNextSession || sinceTheLastTimeOn > sinceTheStartOfTheSession {
		sinceTheStartOfTheSession = 0
	} else {
		sinceTheStartOfTheSession -= sinceTheLastTimeOn
	}
	db.SetDuration(sinceTheStartOfTheSessionKey, sinceTheStartOfTheSession)

	shutdownChannel := make(chan struct{})
	go func() {
		<-shutdownChannel
		Shutdown()
	}()

	for {
		log.Println(time.Since(cycleStart)) //--
		// Is a cycle over?
		if time.Since(cycleStart) >= oneCycle {
			cycleStart = time.Now()
			db.SetTime(cycleStartKey, cycleStart)
			// Reset the usage rate since the start of a cycle.
			db.SetDuration(sinceTheBeginningOfTheCycleKey, 0)
		}

		// Has the usage rate reached the maximum allowed since the beginning of the session?
		sinceTheStartOfTheSession, err := db.GetDuration(sinceTheStartOfTheSessionKey)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("std", sinceTheStartOfTheSession) //--
		if sinceTheStartOfTheSession >= allowedTimeForOneSession {
			shutdownChannel <- struct{}{}
		}

		//Has the usage rate reached the maximum allowed since the beginning of the cycle?
		sinceTheBeginningOfTheCycle, err := db.GetDuration(sinceTheBeginningOfTheCycleKey)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("td", sinceTheBeginningOfTheCycle) //--
		if sinceTheBeginningOfTheCycle >= allowedTimeInOneCycle {
			shutdownChannel <- struct{}{}
		}

		time.Sleep(checkInterval)

		// Increase durations
		db.IncDuration(sinceTheBeginningOfTheCycleKey, checkInterval)
		db.IncDuration(sinceTheStartOfTheSessionKey, checkInterval)

		// Update the last time it was on
		db.SetTime(lastTimeOnKey, time.Now())
	}
}

func Shutdown() {
	log.Println("خاموش شد")
	if err := exec.Command("cmd", "/C", "shutdown", "/s").Run(); err != nil {
		log.Println("Failed to initiate shutdown:", err)
	}
}
