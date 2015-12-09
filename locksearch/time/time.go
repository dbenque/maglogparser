package time

import (
	"fmt"
	"maglogparser/locksearch/record"
	"maglogparser/utils"
	"os"
	"sort"
	"sync"
	"text/tabwriter"
	"time"
)

type result struct {
	indicator string
	tid       string
	r         *record.Record
}

type results struct {
	sync.Mutex
	record.HasRecords
}

func (r *result) GetRecord() *record.Record {
	return r.r
}

func SetTime(tInput time.Time) error {

	var myresults results

	var wg sync.WaitGroup
	for id, sl := range record.RecordByThread {
		wg.Add(1)
		go func(tid string, rs []*record.Record) {
			defer wg.Done()

			for _, r := range rs {
				if r.Time.Equal(tInput) {
					myresults.Lock()
					myresults.HasRecords = append(myresults.HasRecords, &result{indicator: "*", tid: tid, r: r})
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
					myresults.HasRecords = append(myresults.HasRecords, &result{indicator: prefix, tid: tid, r: p})
					myresults.Unlock()
					break
				}

			}

		}(id, sl)
	}

	wg.Wait()

	// output the last log of each TID for that time
	sort.Sort(record.ByHasRecordTime{myresults.HasRecords})

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 1, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Tag\tTID\tRaw Log\t")
	for _, r := range myresults.HasRecords {
		res, _ := r.(*result)
		fmt.Fprintf(w, "%s\t%s\t%s\t\n", res.indicator, res.tid, res.r.Raw)

	}
	fmt.Fprintln(w)
	w.Flush()

	// Search for the ongoing commands
	var inGoingCommandResult results
	for _, r := range myresults.HasRecords {
		res, _ := r.(*result)
		// If we may find a command letś search for it
		if res.indicator != "-" {
			wg.Add(1)
			go func(aRes result) {
				defer wg.Done()
				if r := aRes.r.GetCurrentCommand(); r != nil {
					inGoingCommandResult.Lock()
					aRes.r = r
					inGoingCommandResult.HasRecords = append(inGoingCommandResult.HasRecords, &aRes)
					inGoingCommandResult.Unlock()
				}
			}(*res)
		}
	}

	wg.Wait()

	// output the last Command of each TID for that time
	sort.Sort(record.ByHasRecordTime{inGoingCommandResult.HasRecords})

	wcmd := new(tabwriter.Writer)
	wcmd.Init(os.Stdout, 1, 0, 2, ' ', 0)
	fmt.Fprintln(wcmd, "OnGoing commands at that time:")
	fmt.Fprintln(wcmd, "TID\tCmd\tStart\tRunning for\tRemaining")
	for _, r := range inGoingCommandResult.HasRecords {
		res, _ := r.(*result)
		completionRec := res.r.GetCommandCompletion()
		if completionRec == nil {
			fmt.Fprintf(wcmd, "%s\t%s\t%s\t%f\t--\t\n", res.tid, res.r.Cmd, res.r.Time.Format(utils.DateFormat), tInput.Sub(res.r.Time).Seconds())
			continue
		}
		if completionRec.Time.Before(tInput) {
			continue
		}
		fmt.Fprintf(wcmd, "%s\t%s\t%s\t%f\t%f\t\n", res.tid, res.r.Cmd, res.r.Time.Format(utils.DateFormat), tInput.Sub(res.r.Time).Seconds(), completionRec.Time.Sub(tInput).Seconds())

	}
	fmt.Fprintln(wcmd)
	wcmd.Flush()

	return nil

}
