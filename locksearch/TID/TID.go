package TID

import (
	"fmt"
	"maglogparser/locksearch/record"
	"maglogparser/locksearch/window"
	"os"
	"sort"
	"sync"
	"text/tabwriter"
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

func StatTID() {

	result := statThreadIDCmd()

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 20, 0, 2, ' ', tabwriter.AlignRight)

	fmt.Fprintln(w, "TID\tCommand Count\t")

	for _, stat := range result {
		fmt.Fprintf(w, "%s\t%d\t\n", stat.tid, stat.count)
	}

	fmt.Fprintln(w)
	w.Flush()
}

func statThreadIDCmd() []countTID {

	var result statTIDResults

	var wg sync.WaitGroup
	fmt.Printf("Length:%d\n", len(record.RecordByThread))
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

		}(id, sl)
	}
	wg.Wait()
	sort.Sort(byCount(result.stats))
	return result.stats
}
