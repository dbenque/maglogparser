package queue

import (
	"sort"
	"strconv"
	"strings"
	"sync"

	"maglogparser/locksearch/record"

	tm "github.com/buger/goterm"
)

type result struct {
	qSize int
	tid   string
	r     *record.Record
}

type results struct {
	sync.Mutex
	res []record.HasRecord
}

func (r *result) GetRecord() *record.Record {
	return r.r
}

func StatQueue() error {

	var myresults results
	var wg sync.WaitGroup
	for id, sl := range record.RecordByThread {
		wg.Add(1)
		go func(tid string, rs []*record.Record) {

			defer wg.Done()
			for _, r := range rs {

				if len(r.Raw) < 131 {
					continue
				}
				txt := strings.Split(r.Raw[60:130], ">")
				if len(txt) < 2 {
					continue
				}

				tokens := strings.Split(txt[1], " ") // 60:130 to limit the scope of search, with reasonable margins
				if ((tokens[1] == "Queue") && (tokens[2] == "command")) ||
					((tokens[1] == "Dequeue") && (tokens[2] == "and") && (tokens[3] == "execute")) {
					//2015/08/04 15:19:59.904847 magap302 masterag-11298 MDW INFO <SEI_MAAdminSequence.cpp#223 TID#13> Queue command 0x2aacbf979400 [Command#14015: kSEIBELibLoaded : BE->MAG:Notify the Master Agent that a lib has been loaded (version 1)^@] (from BENT0UCL4VXODY to masterag); Command sequence queue size: 0
					backToken := strings.Split(r.Raw[len(r.Raw)-40:len(r.Raw)], " ")
					last := len(backToken) - 1
					txt1 := strings.Join(backToken[last-2:last], " ")
					if txt1 == "queue size:" {
						if s, err := strconv.Atoi(backToken[last]); err == nil {
							myresults.Lock()
							myresults.res = append(myresults.res, &result{qSize: s, tid: tid, r: r})
							myresults.Unlock()
						}
					}
				}
			}
		}(id, sl)
	}

	wg.Wait()
	sort.Sort(record.ByRecordTime(myresults.res))

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

	t0 := myresults.res[0].(*result).r.Time
	tmax := myresults.res[len(myresults.res)-1].(*result).r.Time
	indexMax := int(tmax.Sub(t0).Seconds())

	m := make(map[int]prec)
	for _, v := range myresults.res {
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
	}

	for t := 0; t <= int(indexMax); t++ {
		data.AddRow(float64(t), float64(m[t].max), float64(m[t].min))
	}

	tm.Println(chart.Draw(data))
	tm.Flush()

	return nil
}
