package cmd

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

type CmdMap struct {
	sync.Mutex
	m map[string][]*record.Record
}

type cmdStat struct {
	cmd        string
	records    []*record.Record
	max        time.Duration
	min        time.Duration
	sum        time.Duration
	incomplete int
}

func (c *cmdStat) GetCmdName() string {
	return c.cmd
}

func (c *cmdStat) GetDuration() (time.Duration, error) {
	return c.GetDuration()
}

func StatCmdTID(aTID string) {
	rs, ok := record.RecordByThread[aTID]
	if !ok {
		fmt.Println("Unknown TID")
		return
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 20, 0, 2, ' ', tabwriter.AlignRight)
	//fmt.Fprintln(w, "Time\tCommand\tEnd\tDuration\t")
	fmt.Fprintln(w, "Time\tCommand\tDuration(s)\tNext Cmd After(s)\t")

	currentCmd := rs[0] // TODO Optim to start at the border of the window

	if currentCmd.PreviousByThread != nil {
		currentCmd = currentCmd.PreviousByThread
	}

	for {
		nextCmd := currentCmd.GetNextCommand()
		if nextCmd == nil {
			break
		}

		if window.InCurrentWindow(nextCmd.Time) {
			nextCmd.GetCommandCompletion()
		}

		currentCmd = nextCmd

		if window.InCurrentWindow(currentCmd.Time) {
			duration := -1.0
			next := -1.0
			if currentCmd.CmdCompletionRecord != nil {
				duration = currentCmd.CmdCompletionRecord.Time.Sub(currentCmd.Time).Seconds()
				nc := currentCmd.GetNextCommand()
				if nc != nil {
					next = nc.Time.Sub(currentCmd.CmdCompletionRecord.Time).Seconds()
				}
			}

			fmt.Fprintf(w, "%s\t%s\t%f\t%f\t\n", currentCmd.Time.Format(utils.DateFormat), currentCmd.Cmd, duration, next)
		}
	}

	fmt.Fprintln(w)
	w.Flush()

}

func StatCmd() {

	result := statCmdPerThread()

	var stats []record.HasCmdName

	var wg sync.WaitGroup
	for cmd, records := range result.m {
		aStat := cmdStat{cmd: cmd, records: records}
		stats = append(stats, &aStat)
		wg.Add(1)
		go func(s *cmdStat) {
			defer wg.Done()
			s.min = 1 * time.Hour
			for _, r := range s.records {
				d, err := r.GetCmdDuration()
				if err != nil {
					s.incomplete++
					continue
				}

				if s.max < d {
					s.max = d
				}
				if s.min > d {
					s.min = d
				}
				s.sum += d
			}

			//sort.Sort(record.ByDuration(stats))

		}(&aStat)
	}

	wg.Wait()

	sort.Sort(record.ByCmdName(stats))

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 20, 0, 2, ' ', tabwriter.AlignRight)

	fmt.Fprintln(w, "Command\tcount\tmiss\tmin (ms)\tmax (ms)\tavg (ms)\t")
	for _, s := range stats {
		ss, _ := s.(*cmdStat)
		c := int64(len(ss.records) - ss.incomplete)
		if c == 0 {
			c = 1
		}
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%d\t\n", ss.cmd, len(ss.records), ss.incomplete, ss.min.Nanoseconds()/1000000, ss.max.Nanoseconds()/1000000, ss.sum.Nanoseconds()/1000000/c)

	}

	fmt.Fprintln(w)
	w.Flush()
}

func statCmdPerThread() CmdMap {

	var result CmdMap
	result.m = make(map[string][]*record.Record)

	var wg sync.WaitGroup
	for id, sl := range record.RecordByThread {

		wg.Add(1)
		go func(tid string, rs []*record.Record) {
			defer wg.Done()

			currentCmd := rs[0] // TODO Optim to start at the border of the window

			if currentCmd.PreviousByThread != nil {
				currentCmd = currentCmd.PreviousByThread
			}

			for {
				nextCmd := currentCmd.GetNextCommand()
				if nextCmd == nil {
					break
				}

				if window.InCurrentWindow(nextCmd.Time) {

					nextCmd.GetCommandCompletion()

					result.Lock()
					records, ok := result.m[nextCmd.Cmd]
					if !ok {
						records = make([]*record.Record, 0, 0)
					}
					result.m[nextCmd.Cmd] = append(records, nextCmd)
					result.Unlock()
				}
				currentCmd = nextCmd
			}

		}(id, sl)
	}
	wg.Wait()
	return result
}
