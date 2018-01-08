package watcher

import (
	"log"
	"os"
	"time"
)

// Watcher controls regular observation
type Watcher struct {
	ChannelsFilePath string
	WatchHour        []int
}

var (
	logger = log.New(os.Stdout, "[watcher]", log.Lshortfile)
)

// nextTimer returns a new timer set at the next hour in WatchHour
func (wtc *Watcher) nextTimer() *time.Timer {
	now := time.Now()
	for _, h := range wtc.WatchHour {
		if now.Hour() < h {
			targetTime := time.Date(now.Year(), now.Month(), now.Day(), h, 0, 0, 0, now.Location())
			return time.NewTimer(targetTime.Sub(now))
		}
	}
	targetTime := time.Date(
		now.Year(), now.Month(), now.Day(), wtc.WatchHour[0], 0, 0, 0, now.Location()).Add(24 * time.Hour) // Tomorrow
	return time.NewTimer(targetTime.Sub(now))
}

// Run executes a given function f according to WatchHour
func (wtc *Watcher) Run(f func(*Watcher)) {
	go func() {
		for {
			t := wtc.nextTimer()
			<-t.C
			f(wtc)
		}
	}()
}

// RunPeriodic executes a given function f periodically
func (wtc *Watcher) RunPeriodic(f func(*Watcher), period time.Duration) {
	go func() {
		t := time.NewTicker(period)
		for {
			<-t.C
			f(wtc)
		}
	}()
}
