package main

import (
	"fmt"
	"maglogparser/locksearch/record"
	"maglogparser/locksearch/window"
	"sort"
	"sync"
)

type countTID struct {
	tid   string
	count int
}

type statTIDResults struct {
	sync.Mutex
	stats []countTID
}

type byCount []countTID

func (a byCount) Len() int           { return len(a) }
func (a byCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byCount) Less(i, j int) bool { return a[i].count < a[j].count }

func statTID() {

	result := statThreadIDCmd()

	for _, stat := range result {
		fmt.Printf("%s : %d\n", stat.tid, stat.count)
	}

}

func statThreadIDCmd() []countTID {

	var result statTIDResults

	var wg sync.WaitGroup
	for id, sl := range record.RecordByThread {
		wg.Add(1)
		go func(tid string, rs []*record.Record) {
			defer wg.Done()

			myStat := countTID{tid: tid, count: 0}
			currentCmd := rs[0] // TODO Optim to start at the border of the window

			if currentCmd.IsCommand() {
				myStat.count = 1
			}

			for {
				nextCmd := currentCmd.GetNextCommand()
				if nextCmd == nil {
					break
				}

				if window.InCurrentWindow(nextCmd.Time) {
					myStat.count++
				}

				currentCmd = nextCmd
			}

			result.Lock()
			result.stats = append(result.stats, myStat)
			result.Unlock()

			// max := time.Duration(0)
			// var rmax *record.Record
			// for _, r := range rs {
			//
			// 	if !window.InCurrentWindow(r.Time) {
			// 		continue
			// 	}
			//
			// 	if r.NextByThread != nil {
			// 		d := r.NextByThread.Time.Sub(r.Time)
			// 		if d > max {
			// 			max = d
			// 			rmax = r
			// 		}
			// 	}
			// }
			//
			// if rmax != nil {
			// 	rcmd := rmax.GetCurrentCommand()
			//
			// 	rcmdCompleted := rmax.GetCommandCompletion()
			// 	result.Lock()
			// 	result.m = append(result.m, statRecord{tid: tid, r: rmax, rcmd: rcmd, rcmdCompleted: rcmdCompleted})
			// 	result.Unlock()
			// }

		}(id, sl)
	}
	wg.Wait()
	sort.Sort(byCount(result.stats))
	return result.stats
}
