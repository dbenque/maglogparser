package window

import (
	"fmt"
	"maglogparser/locksearch/record"
	"maglogparser/utils"
	"sync"
	"time"
)

type Window struct {
	sync.Mutex
	StartWindow time.Time
	EndWindow   time.Time

	Sooner time.Time
	Later  time.Time

	Current time.Time
}

func (w Window) Print() string {
	return fmt.Sprintf("[ %s ] / %s / [ %s ]", w.Sooner.Format(utils.DateFormat), w.PrintCurrent(), w.Later.Format(utils.DateFormat))
}

func (w Window) PrintCurrent() string {
	return fmt.Sprintf("%s  <--->  %s", w.StartWindow.Format(utils.DateFormat), w.EndWindow.Format(utils.DateFormat))
}

var window Window

func GetWindow() Window {
	w := window
	return w
}

func SetTime(t time.Time) error {

	if t.After(window.Later) && t.Before(window.Sooner) {
		return fmt.Errorf("Time out of max window range.")
	}

	window.Current = t

	return nil
}

func Reset() {
	window.Lock()
	defer window.Unlock()

	window.StartWindow = window.Sooner
	window.EndWindow = window.Later
}

func SetStart(t time.Time) error {
	window.Lock()
	defer window.Unlock()

	if t.After(window.EndWindow) {
		return fmt.Errorf("Cannot set start after end.")
	}

	if t.Before(window.Sooner) {
		return fmt.Errorf("Cannot set start before %s", window.Sooner.Format(utils.DateFormat))
	}

	window.StartWindow = t

	return nil
}

func SetEnd(t time.Time) error {
	window.Lock()
	defer window.Unlock()

	if t.Before(window.StartWindow) {
		return fmt.Errorf("Cannot set end before start.")
	}

	if t.After(window.Later) {
		return fmt.Errorf("Cannot set end after %s", window.Later.Format(utils.DateFormat))
	}

	window.EndWindow = t

	return nil
}

func InCurrentWindow(t time.Time) bool {
	return t.After(window.StartWindow) && t.Before(window.EndWindow)
}

func injectNewEvent(t time.Time) {
	window.Lock()
	defer window.Unlock()

	if window.StartWindow.IsZero() || t.Before(window.StartWindow) {
		window.StartWindow = t
		window.Sooner = t
	}

	if window.EndWindow.IsZero() || t.After(window.EndWindow) {
		window.EndWindow = t
		window.Later = t
	}
}

// ScanThreadID place the received record in the good map per TID and chain the record previous/next
func ScanTime(c <-chan *record.Record) {

	for r := range c {
		injectNewEvent(r.Time)
	}
}
