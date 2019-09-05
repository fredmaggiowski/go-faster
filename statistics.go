package main

import (
	"time"
)

type TimeTrack struct {
	Start time.Time
	End   time.Time
}

// Statistics is used to log game statistics.
type Statistics struct {
	Speed map[string]TimeTrack
}

func newStatistics() *Statistics {
	return &Statistics{
		Speed: make(map[string]TimeTrack),
	}
}

func (s *Statistics) getWinner() string {
	var winner string
	deltaWinner, _ := time.ParseDuration("2000h")

	for player, track := range s.Speed {
		delta := track.End.Sub(track.Start)
		if delta < deltaWinner {
			deltaWinner = delta
			winner = player
		}
	}

	return winner
}
