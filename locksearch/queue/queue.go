package queue

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"maglogparser/locksearch/record"
	"maglogparser/locksearch/window"
	"maglogparser/utils"

	tm "github.com/buger/goterm"
)

type result struct {
	qSize int
	tid   string
	r     *record.Record
}

type results struct {
	sync.Mutex
	record.HasRecords
}

func (r *result) GetRecord() *record.Record {
	return r.r
}

func StatQueue(writeQ bool) error {

	var myresults results
	var wg sync.WaitGroup
	for id, sl := range record.RecordByThread {
		wg.Add(1)
		go func(tid string, rs []*record.Record) {

			defer wg.Done()

			for _, r := range rs {

				if !window.InCurrentWindow(r.Time) {
					continue
				}

				if len(r.Raw) < 151 {
					continue
				}
				txt := strings.Split(r.Raw[60:150], ">")
				if len(txt) < 2 {
					continue
				}
				tokens := strings.Split(txt[1], " ") // 60:140 to limit the scope of search, with reasonable margins
				if len(tokens) < 3 {
					continue
				}

				if writeQ {
					if ((tokens[1] == "Queue") && (tokens[2] == "command")) ||
						((tokens[1] == "Dequeue") && (tokens[2] == "and") && (tokens[3] == "execute")) {
						//2015/08/04 15:19:59.904847 magap302 masterag-11298 MDW INFO <SEI_MAAdminSequence.cpp#223 TID#13> Queue command 0x2aacbf979400 [Command#14015: kSEIBELibLoaded : BE->MAG:Notify the Master Agent that a lib has been loaded (version 1)^@] (from BENT0UCL4VXODY to masterag); Command sequence queue size: 0
						backToken := strings.Split(r.Raw[len(r.Raw)-40:len(r.Raw)], " ")
						last := len(backToken) - 1
						txt1 := strings.Join(backToken[last-2:last], " ")
						if txt1 == "queue size:" {
							if s, err := strconv.Atoi(backToken[last]); err == nil {
								myresults.Lock()
								myresults.HasRecords = append(myresults.HasRecords, &result{qSize: s, tid: tid, r: r})
								myresults.Unlock()
							}
						}
					}
				} else {
					if (tokens[1] == "Scheduling") && (tokens[3] == "command") && (tokens[6] == "Read-only") {
						//2015/12/03 05:07:12.577382 magap302 masterag-32357 MDW INFO <SEI_MAAdminSequence.cpp#196 TID#13> Scheduling the command in the Read-only command sequencer pool. Pool queue size: 393
						backToken := strings.Split(r.Raw[len(r.Raw)-40:len(r.Raw)], " ")
						last := len(backToken) - 1
						txt1 := strings.Join(backToken[last-2:last], " ")
						if txt1 == "queue size:" {
							if s, err := strconv.Atoi(backToken[last]); err == nil {
								myresults.Lock()
								myresults.HasRecords = append(myresults.HasRecords, &result{qSize: s, tid: tid, r: r})
								myresults.Unlock()
							}
						}
					}
				}
			}
		}(id, sl)
	}

	wg.Wait()
	sort.Sort(record.ByHasRecordTime{myresults.HasRecords})

	// Build chart
	chart := tm.NewLineChart(tm.Width()-10, tm.Height()-10)

	data := new(tm.DataTable)
	data.AddColumn("Seconds")
	data.AddColumn("QMax")
	data.AddColumn("QMin")

	type prec struct {
		max int
		min int
	}

	t0 := myresults.HasRecords[0].(*result).r.Time
	tmax := myresults.HasRecords[len(myresults.HasRecords)-1].(*result).r.Time
	indexMax := int(tmax.Sub(t0).Seconds())

	qmaxTime := time.Unix(0, 0)
	qmax := 0

	m := make(map[int]prec)
	for _, v := range myresults.HasRecords {
		r := v.(*result)
		d := int(r.r.Time.Sub(t0).Seconds())
		p, ok := m[d]
		if !ok {
			p = prec{r.qSize, r.qSize}
		} else {
			if p.max < r.qSize {
				p.max = r.qSize
			}
			if p.min > r.qSize {
				p.min = r.qSize
			}
		}
		m[d] = p

		if p.max > qmax {
			qmax = p.max
			qmaxTime = r.GetRecord().Time
		}
	}

	for t := 0; t <= int(indexMax); t++ {
		data.AddRow(float64(t), float64(m[t].max), float64(m[t].min))
	}

	tm.Println(chart.Draw(data))
	tm.Flush()

	fmt.Printf("Qmax=%d  at %s\n", qmax, qmaxTime.Format(utils.DateFormat))

	return nil
}
