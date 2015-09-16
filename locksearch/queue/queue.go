package queue

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"maglogparser/locksearch/record"
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
	fmt.Printf("Length:%d\n", len(record.RecordByThread))
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
				if (tokens[1] == "Queue") && (tokens[2] == "command") {

					//2015/08/04 15:19:59.904847 magap302 masterag-11298 MDW INFO <SEI_MAAdminSequence.cpp#223 TID#13> Queue command 0x2aacbf979400 [Command#14015: kSEIBELibLoaded : BE->MAG:Notify the Master Agent that a lib has been loaded (version 1)^@] (from BENT0UCL4VXODY to masterag); Command sequence queue size: 0
					backToken := strings.Split(r.Raw[len(r.Raw)-40:len(r.Raw)], " ")
					last := len(backToken) - 1
					txt1 := strings.Join(backToken[last-4:last], " ")
					if txt1 == "Command sequence queue size:" {

						if s, err := strconv.Atoi(backToken[last]); err == nil {
							myresults.Lock()
							myresults.res = append(myresults.res, &result{qSize: s, tid: tid, r: r})
							myresults.Unlock()
						} else {
							fmt.Printf("err #%s#\n", r.Raw)
						}
					}

				}

			}

		}(id, sl)
	}

	wg.Wait()
	fmt.Println("!! %d", len(myresults.res))
	sort.Sort(record.ByRecordTime(myresults.res))
	fmt.Println("!! %d", len(myresults.res))
	//
	// w := new(tabwriter.Writer)
	// w.Init(os.Stdout, 1, 0, 2, ' ', 0)
	// fmt.Fprintln(w, "Tag\tTID\tRaw Log\t")
	// for _, r := range myresults.res {
	// 	res, _ := r.(*result)
	// 	fmt.Fprintf(w, "%s\t%s\t%s\t\n", res.indicator, res.tid, res.r.Raw)
	//
	// }
	// fmt.Fprintln(w)
	// w.Flush()

	return nil
}
