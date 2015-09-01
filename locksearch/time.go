package main

import (
	"fmt"
	"maglogparser/locksearch/record"
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

	sort.Sort(ByRecordTime(myresults.res))

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 1, 0, 2, ' ', 0)

	fmt.Fprintln(w, "Tag\tTID\tRaw Log\t")

	for _, r := range myresults.res {
		rec, _ := r.(*result)
		fmt.Fprintf(w, "%s\t%s\t%s\t\n", rec.indicator, rec.tid, rec.r.Raw)
	}
	fmt.Fprintln(w)
	w.Flush()
	return nil

}
