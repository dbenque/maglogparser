package main

import (
	"fmt"
	"maglogparser/locksearch/record"
	"maglogparser/utils"
	"os"
	"text/tabwriter"
	"time"
)
import (
	"sort"
	"sync"
)

type result struct {
	indicator string
	tid       string
	r         *record.Record
}

type results struct {
	sync.Mutex
	res []HasRecord
}

func (r *result) GetRecord() *record.Record {
	return r.r
}

func setTime(tInput time.Time) error {

	var myresults results

	var wg sync.WaitGroup
	for id, sl := range record.RecordByThread {
		wg.Add(1)
		go func(tid string, rs []*record.Record) {
			defer wg.Done()

			for _, r := range rs {
				if r.Time.Equal(tInput) {
					myresults.Lock()
					myresults.res = append(myresults.res, &result{indicator: "*", tid: tid, r: r})
					myresults.Unlock()
					break
				}

				if r.Time.After(tInput) {
					myresults.Lock()
					p := r.PreviousByThread
					prefix := ""
					if p == nil {
						p = r
						prefix = "-"
					}
					myresults.res = append(myresults.res, &result{indicator: prefix, tid: tid, r: p})
					myresults.Unlock()
					break
				}

			}

		}(id, sl)
	}

	wg.Wait()

	// output the last log of each TID for that time
	sort.Sort(ByRecordTime(myresults.res))

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 1, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Tag\tTID\tRaw Log\t")
	for _, r := range myresults.res {
		res, _ := r.(*result)
		fmt.Fprintf(w, "%s\t%s\t%s\t\n", res.indicator, res.tid, res.r.Raw)

	}
	fmt.Fprintln(w)
	w.Flush()

	// Search for the ongoing commands
	var inGoingCommandResult results
	for _, r := range myresults.res {
		res, _ := r.(*result)
		// If we may find a command letś search for it
		if res.indicator != "-" {
			wg.Add(1)
			go func(aRes result) {
				defer wg.Done()
				if r := aRes.r.GetCurrentCommand(); r != nil {
					inGoingCommandResult.Lock()
					aRes.r = r
					inGoingCommandResult.res = append(inGoingCommandResult.res, &aRes)
					inGoingCommandResult.Unlock()
				}
			}(*res)
		}
	}

	wg.Wait()

	// output the last Command of each TID for that time
	sort.Sort(ByRecordTime(inGoingCommandResult.res))

	wcmd := new(tabwriter.Writer)
	wcmd.Init(os.Stdout, 1, 0, 2, ' ', 0)
	fmt.Fprintln(wcmd, "OnGoing commands at that time:")
	fmt.Fprintln(wcmd, "TID\tCmd\tStart\tRunning for\t")
	for _, r := range inGoingCommandResult.res {
		res, _ := r.(*result)
		fmt.Fprintf(wcmd, "%s\t%s\t%s\t%f\t\n", res.tid, res.r.Cmd, res.r.Time.Format(utils.DateFormat), tInput.Sub(res.r.Time).Seconds())

	}
	fmt.Fprintln(wcmd)
	wcmd.Flush()

	return nil

}
