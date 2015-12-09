package cmd

import (
	"fmt"
	"maglogparser/locksearch/record"
	"maglogparser/locksearch/window"
	"maglogparser/utils"
	"os"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	tm "github.com/buger/goterm"
)

type cmdMap struct {
	sync.Mutex
	m map[string]record.Records
}

type cmdStat struct {
	cmd        string
	records    record.Records
	max        time.Duration
	min        time.Duration
	sum        time.Duration
	incomplete int
}

type cmdStats []*cmdStat

// To sort stat by command name
func (a cmdStats) Len() int      { return len(a) }
func (a cmdStats) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a cmdStats) Less(i, j int) bool {
	return strings.Compare(a[i].cmd, a[j].cmd) < 0
	//return a[i].GetCmdName()[0] < a[j].GetCmdName()[0]
}

// Write all the commands of the given thread
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

// Build statistics for all commands
func buildStats() cmdStats {
	result := statCmdPerThread()

	stats := cmdStats{}

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

			sort.Sort(record.ByDuration{s.records})

		}(&aStat)
	}

	wg.Wait()

	sort.Sort(stats)

	return stats
}

//Display the distribution for a given command
func StatCmdDistribution(cmdName string) {

	for _, ss := range buildStats() {
		if ss.cmd == cmdName {
			// Build chart
			chart := tm.NewLineChart(tm.Width()-10, tm.Height()-33)

			data := new(tm.DataTable)
			data.AddColumn("CmdCount")
			data.AddColumn("Duration")

			for i, rec := range ss.records {
				if d, err := rec.GetCmdDuration(); err != nil {
					data.AddRow(float64(i), -1.)
				} else {
					data.AddRow(float64(i), float64(d.Nanoseconds()/1000000))
				}
			}

			tm.Println(chart.Draw(data))
			tm.Flush()

			// Statistics for that particular command
			{
				w := new(tabwriter.Writer)
				w.Init(os.Stdout, 20, 0, 2, ' ', tabwriter.AlignRight)

				fmt.Fprintln(w, "Command\tcount\tmiss\tmin (ms)\tmax (ms)\tavg (ms)\t95% (ms)\t")
				c := int64(len(ss.records) - ss.incomplete)
				i95 := c * 95 / 100

				if c == 0 {
					c = 1
				}
				percentile95 := int64(-1)
				if d95, err := ss.records[i95].GetCmdDuration(); err == nil {
					percentile95 = d95.Nanoseconds()
				}

				fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%d\t%d\t\n", ss.cmd, len(ss.records), ss.incomplete, ss.min.Nanoseconds()/1000000, ss.max.Nanoseconds()/1000000, ss.sum.Nanoseconds()/1000000/c, percentile95/1000000)

				fmt.Fprintln(w)
				w.Flush()

			}
			// The 10 slowest occurence
			{
				w := new(tabwriter.Writer)
				w.Init(os.Stdout, 5, 0, 2, ' ', tabwriter.TabIndent)
				fmt.Fprintln(w, "Duration(ms)\tcmd\t")

				min := 10
				if len(ss.records) < 10 {
					min = len(ss.records)
				}
				for i := 1; i <= min; i++ {
					rec := ss.records[len(ss.records)-i]
					if d, err := rec.GetCmdDuration(); err != nil {
						fmt.Fprintf(w, "-\t%s\t\n", rec.Raw)
					} else {
						fmt.Fprintf(w, "%d\t%s\t\n", d.Nanoseconds()/1000000, rec.Raw)
					}

				}

				fmt.Fprintln(w)
				w.Flush()
			}
			return
		}
	}

}

// Display all the statistics
func StatCmdAll() {

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 20, 0, 2, ' ', tabwriter.AlignRight)

	fmt.Fprintln(w, "Command\tcount\tmiss\tmin (ms)\tmax (ms)\tavg (ms)\t95% (ms)\t")
	for _, ss := range buildStats() {
		c := int64(len(ss.records) - ss.incomplete)
		i95 := c * 95 / 100

		if c == 0 {
			c = 1
		}
		percentile95 := int64(-1)
		if d95, err := ss.records[i95].GetCmdDuration(); err == nil {
			percentile95 = d95.Nanoseconds()
		}

		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%d\t%d\t\n", ss.cmd, len(ss.records), ss.incomplete, ss.min.Nanoseconds()/1000000, ss.max.Nanoseconds()/1000000, ss.sum.Nanoseconds()/1000000/c, percentile95/1000000)
	}

	fmt.Fprintln(w)
	w.Flush()
}

func statCmdPerThread() cmdMap {

	var result cmdMap
	result.m = make(map[string]record.Records)

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
