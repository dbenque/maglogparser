package lock

import (
	"fmt"
	"maglogparser/locksearch/record"
	"maglogparser/locksearch/window"
	"maglogparser/utils"
	"os"
	"sort"
	"sync"
	"text/tabwriter"
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

func StatLock(onlyCmd bool, setWindow bool) {

	result := statThreadID()

	start := time.Unix(0, 0)
	end := time.Now().Add(24 * time.Hour)

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 20, 0, 2, ' ', tabwriter.AlignRight)

	fmt.Fprintln(w, "Time\tTID\tWait\tTill\tCommand\tStart\tEnd\tDuration\t")

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

		if stat.r.Time.After(start) {
			start = stat.r.Time
		}

		if stat.r.NextByThread.Time.After(end) {
			end = stat.r.NextByThread.Time
		}

		fmt.Fprintf(w, "%s\t%s\t%f\t%s\t%s\t%s\t%s\t%s\t\n", stat.r.Time.Format(utils.DateFormat), stat.tid, max.Seconds(), stat.r.NextByThread.Time.Format(utils.DateFormat), cmd, cmdTime, cmdCompleteTime, duration)

		if setWindow {
			window.Reset()
			window.SetStart(start)
			window.SetEnd(end)
		}

	}
	fmt.Fprintln(w)
	w.Flush()

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
				rcmdCompleted := rcmd
				if rcmd != nil {
					rcmdCompleted = rmax.GetCommandCompletion()
				}
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
