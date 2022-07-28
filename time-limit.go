package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/nathan-osman/go-sunrise"
	"github.com/smorz/time-limit/database"
)

const (
	constCheckInterval                    = time.Minute
	constAllowedTimeForOneCycle           = time.Minute * 150
	contsOneCycle                         = 14*time.Hour + 35*time.Minute
	constAllowedTimeForOneSession         = time.Minute * 50
	constNecessaryRestUntilTheNextSession = time.Minute * 30
	Latitude                              = 35.6892
	Longitude                             = 51.3890

	cycleStartKey                  = "cycle_start"
	sinceTheBeginningOfTheCycleKey = "since_the_beginning_of_the_cycle"
	lastTimeOnKey                  = "last_time_on"
	sinceTheStartOfTheSessionKey   = "since_the_start_of_the_session"
)

var (
	checkInterval                    = constCheckInterval
	allowedTimeForOneCycle           = constAllowedTimeForOneCycle
	oneCycle                         = contsOneCycle
	allowedTimeForOneSession         = constAllowedTimeForOneSession
	necessaryRestUntilTheNextSession = constNecessaryRestUntilTheNextSession
	debug                            = false
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file")
	}
	if allowedCycleMin := os.Getenv("allowed_time_for_one_cycle_min"); allowedCycleMin != "" {
		acm, err := strconv.ParseInt(allowedCycleMin, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		allowedTimeForOneCycle = time.Duration(acm) * time.Minute
	}
	if allowedSessionMin := os.Getenv("allowed_time_for_one_session_min"); allowedSessionMin != "" {
		asm, err := strconv.ParseInt(allowedSessionMin, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		allowedTimeForOneSession = time.Duration(asm) * time.Minute
	}
	if os.Getenv("debug") == "true" {
		checkInterval /= 3000
		allowedTimeForOneCycle /= 3000
		oneCycle = time.Minute
		allowedTimeForOneSession /= 3000
		necessaryRestUntilTheNextSession = time.Second * 10
		debug = true
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if !debug {
		logFile, err := os.OpenFile("time-limit.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(logFile)
	}
	db, err := database.OpenDB("data")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		db.Close()
		Shutdown()
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, exit := context.WithCancel(context.Background())
	go func() {
		oscall := <-c
		log.Printf("system call: %+v\n", oscall)
		exit()
	}()

	cycleStart, err := db.GetTime(cycleStartKey)
	if err != nil {
		log.Fatal(err)
	}
	if debug {
		fmt.Printf("Cycle Start: %v\n", cycleStart)
	}

	check := time.NewTicker(checkInterval)
	var restarted bool

	for {
		if IsNight() {
			log.Println("ّIt is night!")
			return
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
			db.Set(sinceTheStartOfTheSessionKey, sinceTheStartOfTheSession)
		}
		// Is a cycle over?
		if time.Since(cycleStart) >= oneCycle {
			cycleStart = time.Now()
			db.Set(cycleStartKey, cycleStart)
			// Reset the usage rate since the start of a cycle.
			db.Set(sinceTheBeginningOfTheCycleKey, 0)
			if debug {
				fmt.Printf("Cycle start has been reset to: %v\n", cycleStart)
			}
		}

		// Has the usage rate reached the maximum allowed since the beginning of the session?
		if sinceTheStartOfTheSession >= allowedTimeForOneSession {
			log.Println("Reached the maximum time allowed for one session.")
			return
		}

		//Has the usage rate reached the maximum allowed since the beginning of the cycle?
		sinceTheBeginningOfTheCycle, err := db.GetDuration(sinceTheBeginningOfTheCycleKey)
		if err != nil {
			log.Fatal(err)
		}
		if sinceTheBeginningOfTheCycle >= allowedTimeForOneCycle {
			log.Println("Reached the maximum time allowed for one cycle.")
			return
		}
		if debug {
			fmt.Printf("sbs: %v, sbc: %v, now: %v\n", sinceTheStartOfTheSession, sinceTheBeginningOfTheCycle, time.Now())

		}

		select {
		case <-ctx.Done():
			return
		case <-check.C:
		}

		// Increase durations
		db.IncDuration(sinceTheBeginningOfTheCycleKey, checkInterval)
		db.IncDuration(sinceTheStartOfTheSessionKey, checkInterval)

		// Update the last time it was on
		if restarted || time.Since(lastTimeOn) < checkInterval*2 {
			db.Set(lastTimeOnKey, time.Now())
			restarted = false
		}

	}
}

func IsNight() bool {
	now := time.Now()
	r, s := sunrise.SunriseSunset(Latitude, Longitude, now.Year(), now.Month(), now.Day())
	return !(now.After(r.Local()) && now.Before(s.Local()))
}

func Shutdown() {
	log.Println("Shutdown")
	if debug {
		return
	}
	if err := exec.Command("cmd", "/C", "shutdown", "/s").Run(); err != nil {
		log.Println("Failed to initiate shutdown:", err)
	}
}
