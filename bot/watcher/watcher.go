package wacther

import (
	"log"
	"os"
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
func (w *Watcher) nextTimer() time.Timer {
	now := time.Now()
	for _, h := range WatchHour {
		if now.Hour() < h {
			targetTime := time.Date(now.Year(), now.Month(), now.Day(), h)
			return time.NewTimer(targetTime.Sub(now))
		}
	}
	targetTime := time.Data(now.Year(), now.Month(), now.Day(), h).Add(24 * time.Hour) // Tomorrow
	return time.NewTimer(targetTime.Sub(now))
}

// Run executes a given function f according to WatchHour
func (w *Watcher) Run(f func()) {
	go func() {
		for {
			t := w.nextTimer()
			<-t.C
			f()
		}
	}()
}

// Run executes a given function f periodically
func (w *Watcher) RunPeriodic(f func(*Watcher), period time.Duration) {
	go func() {
		t := time.NewTicker(period)
		for {
			<-t.C
			f(w)
		}
	}()
}
