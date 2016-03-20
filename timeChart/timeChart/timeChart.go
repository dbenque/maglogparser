package timeChart

import (
	"fmt"
	"sort"
	"sync"
	"time"

	tm "github.com/buger/goterm"
)

var (
	MaxTimeSlot = 500
)

type TimeData struct {
	When   time.Time
	Series []float64
}

type TimeDataAggregated struct {
	Max   float64
	Min   float64
	Avg   float64
	Count int
}

type TimeDatas []*TimeData

func (a TimeDatas) Len() int {
	return len(a)
}

func (a TimeDatas) Less(i, j int) bool {
	return a[i].When.Before(a[j].When)
}

func (a TimeDatas) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

type TimeChart struct {
	t0, tmax            time.Time            // Internally computed to get time boundaries
	periodSeconds       int                  // period of aggregation, given as parameter, and recomputed if give birth to more data than MaxTimeSlot
	indexMax            int                  // effective number of data displayed on time axis
	tdChan              chan *TimeData       // collection of data
	allData             TimeDatas            // raw data
	results             []TimeDataAggregated // aggregated data
	wgCollect, wgBuild  sync.WaitGroup       // sync between go routine for collection, build and draw steps
	timeFormat          string               // format used to display time on time axis
	noMax, noMin, noAvg bool                 // what are the curves to be displayed
}

func (tc *TimeChart) Add(td *TimeData) {
	tc.tdChan <- td
}

func (tc *TimeChart) DataCompleted() {
	close(tc.tdChan)
}

func (tc *TimeChart) Draw() {

	tc.wgBuild.Wait()
	// Build chart
	chart := tm.NewLineChart(tm.Width()-10, tm.Height()-10)
	dataTable := new(tm.DataTable)

	cparam := 1
	if !tc.noMax {
		cparam++
	}
	if !tc.noMin {
		cparam++
	}
	if !tc.noAvg {
		cparam++
	}

	if cparam == 1 {
		tc.noMax = false
		cparam++
	}

	dataTable.AddColumn(fmt.Sprintf("From %s to %s using Block of %d seconds", tc.t0.Format(tc.timeFormat), tc.tmax.Format(tc.timeFormat), tc.periodSeconds))
	if !tc.noMax {
		dataTable.AddColumn("QMax")
	}
	if !tc.noMin {
		dataTable.AddColumn("QMin")
	}
	if !tc.noAvg {
		dataTable.AddColumn("QAvg")
	}

	for t := 0; t <= int(tc.indexMax); t++ {
		if tc.results[t].Count != 0 {
			ds := []float64{float64(t)}
			if !tc.noMax {
				ds = append(ds, tc.results[t].Max)
			}
			if !tc.noMin {
				ds = append(ds, tc.results[t].Min)
			}
			if !tc.noAvg {
				ds = append(ds, tc.results[t].Avg)
			}

			if cparam == 4 {
				dataTable.AddRow(ds[0], ds[1], ds[2], ds[3])
			}
			if cparam == 3 {
				dataTable.AddRow(ds[0], ds[1], ds[2])
			}
			if cparam == 2 {
				dataTable.AddRow(ds[0], ds[1])
			}
		} else {
			if cparam == 4 {
				dataTable.AddRow(float64(t), 0, 0, 0)
			}
			if cparam == 3 {
				dataTable.AddRow(float64(t), 0, 0)
			}
			if cparam == 2 {
				dataTable.AddRow(float64(t), 0)
			}
		}
	}

	tm.Println(chart.Draw(dataTable))
	tm.Flush()

}

func MakeChart(periodSeconds int, timeFormat string, max, min, avg bool) *TimeChart {

	tc := TimeChart{periodSeconds: periodSeconds, timeFormat: timeFormat, noMax: !max, noMin: !min, noAvg: !avg}
	tc.tdChan = make(chan *TimeData, MaxTimeSlot)
	tc.allData = TimeDatas{}

	tc.wgCollect.Add(1)
	tc.wgBuild.Add(1)

	go func() {
		defer tc.wgCollect.Done()
		for td := range tc.tdChan {
			tc.allData = append(tc.allData, td)
		}
	}()

	go func() {
		tc.wgCollect.Wait()
		defer tc.wgBuild.Done()

		sort.Sort(tc.allData)

		tc.t0 = tc.allData[0].When
		tc.tmax = tc.allData[len(tc.allData)-1].When

		tc.indexMax = 0
		for {
			tc.indexMax = (int(tc.tmax.Sub(tc.t0).Seconds()) / tc.periodSeconds)
			if tc.indexMax > 250 {
				tc.periodSeconds = tc.periodSeconds * 10
				fmt.Printf("Adjust Period: %ds\n", tc.periodSeconds)
			} else {
				break
			}
		}
		tc.results = make([]TimeDataAggregated, tc.indexMax+1)

		for _, d := range tc.allData {
			index := int(d.When.Sub(tc.t0).Seconds()) / tc.periodSeconds
			if tc.results[index].Count == 0 {
				tc.results[index].Max = d.Series[0]
				tc.results[index].Min = d.Series[0]
				tc.results[index].Avg = d.Series[0]
				tc.results[index].Count = 1
			} else {
				if d.Series[0] > tc.results[index].Max {
					tc.results[index].Max = d.Series[0]
				}
				if d.Series[0] < tc.results[index].Min {
					tc.results[index].Min = d.Series[0]
				}
				tc.results[index].Avg = (d.Series[0] + tc.results[index].Avg*float64(tc.results[index].Count)) / float64(tc.results[index].Count+1)
				tc.results[index].Count = tc.results[index].Count + 1
			}
		}
	}()
	return &tc
}
