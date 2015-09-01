package main

import (
	"fmt"
	"maglogparser/locksearch/record"
	"maglogparser/locksearch/window"
	"maglogparser/utils"
	"sort"
	"sync"
	"time"
)

type statRecord struct {
	tid           string
	r             *record.Record
	rcmd          *record.Record
	rcmdCompleted *record.Record
}

type byTime []statRecord

func (a byTime) Len() int           { return len(a) }
func (a byTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byTime) Less(i, j int) bool { return a[i].r.Time.Before(a[j].r.Time) }

type ResultMap struct {
	sync.Mutex
	m []statRecord
}

func statLock(onlyCmd bool) {

	result := statThreadID()

	for _, stat := range result {

		if onlyCmd && stat.rcmd == nil {
			continue
		}

		max := stat.r.NextByThread.Time.Sub(stat.r.Time)

		cmd := "nil"
		cmdTime := "nil"
		if stat.rcmd != nil {
			cmdTime = stat.rcmd.Time.Format(utils.DateFormat)
			cmd = stat.rcmd.Cmd
		}

		cmdCompleteTime := "nil"
		if stat.rcmdCompleted != nil {
			cmdCompleteTime = stat.rcmdCompleted.Time.Format(utils.DateFormat)
		}

		duration := ""
		if stat.rcmdCompleted != nil && stat.rcmd != nil {
			duration = fmt.Sprintf("%f", stat.rcmdCompleted.Time.Sub(stat.rcmd.Time).Seconds())
		}

		fmt.Printf("%s --> TID%s : wait %f till %s for %s since %s to finish at %s duration %s\n", stat.r.Time.Format(utils.DateFormat), stat.tid, max.Seconds(), stat.r.NextByThread.Time.Format(utils.DateFormat), cmd, cmdTime, cmdCompleteTime, duration)

	}

}

func statThreadID() []statRecord {

	var result ResultMap

	var wg sync.WaitGroup
	for id, sl := range record.RecordByThread {
		wg.Add(1)
		go func(tid string, rs []*record.Record) {
			defer wg.Done()
			max := time.Duration(0)
			var rmax *record.Record
			for _, r := range rs {

				if !window.InCurrentWindow(r.Time) {
					continue
				}

				if r.NextByThread != nil {
					d := r.NextByThread.Time.Sub(r.Time)
					if d > max {
						max = d
						rmax = r
					}
				}
			}

			if rmax != nil {
				rcmd := rmax.GetCurrentCommand()

				rcmdCompleted := rmax.GetCommandCompletion()
				result.Lock()
				result.m = append(result.m, statRecord{tid: tid, r: rmax, rcmd: rcmd, rcmdCompleted: rcmdCompleted})
				result.Unlock()
			}

		}(id, sl)
	}
	wg.Wait()
	sort.Sort(byTime(result.m))
	return result.m

}
