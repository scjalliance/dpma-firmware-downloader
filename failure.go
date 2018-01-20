package main

import (
	"log"
	"os"
	"time"
)

func failed(attempt, max int, action string, subaction string, err error) bool {
	if err == nil {
		return false
	}

	if subaction == "" {
		subaction = action
	}

	log.Printf("Unable to %s: %v", subaction, err)

	if attempt+1 >= max {
		log.Printf("The dpma firmware downloader failed after %d attempts to %s.", max, action)
		os.Exit(2)
	}

	d := doublingBackoff(delay, attempt+1)

	log.Printf("Waiting %v then trying again for attempt %d of %d...", d, attempt+2, max)
	time.Sleep(d)

	return true
}

func doublingBackoff(amount time.Duration, rounds int) (backoff time.Duration) {
	if rounds < 2 {
		return amount
	}
	return time.Duration(int64(amount) << uint(rounds-1))
}
