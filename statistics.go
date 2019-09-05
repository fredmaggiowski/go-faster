package main

// TimeTrack is used to log start and end of a player turn.
type TimeTrack struct {
	Start int
	End   int
}

// Statistics is used to log game statistics.
type Statistics struct {
	StartTimings map[string]int
	EndTimings   map[string]int
}
