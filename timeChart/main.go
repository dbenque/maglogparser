package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dbenque/maglogparser/utils"

	tm "github.com/buger/goterm"
)

// var timeFormatFlag = flag.String("t", utils.DateFormat, "Time Format")
// var preriodUnitFlag = flag.String("u", "second", "period [second|minute|hour|day]")
// var preriodValueFlag = flag.Uint("p", 1, "Value of the period")

var timeFormatFlag string
var preriodUnitFlag string
var preriodValueFlag uint
var serieCountFlag uint
var fileFlag string
var noMaxFlag bool
var noMinFlag bool
var noAvgFlag bool

func init() {
	flag.StringVar(&timeFormatFlag, "t", utils.DateFormat, "Time Format")
	flag.StringVar(&preriodUnitFlag, "u", "s", "Unit of the period")
	flag.UintVar(&preriodValueFlag, "p", 1, "Value of the period")
	flag.UintVar(&serieCountFlag, "c", 1, "Number of series")
	flag.StringVar(&fileFlag, "f", "", "file to read")
	flag.BoolVar(&noMaxFlag, "noMax", false, "Don't display Max")
	flag.BoolVar(&noMaxFlag, "noMin", false, "Don't display Min")
	flag.BoolVar(&noAvgFlag, "noAvg", false, "Don't display Avg")
}

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

func main() {

	flag.Parse()
	fmt.Printf("Time format: %s\nPeriod: %d%s\n", timeFormatFlag, preriodValueFlag, preriodUnitFlag)

	if len(fileFlag) == 0 {
		fmt.Println("file input mandatory")
		return
	}

	freader, err := os.Open(fileFlag)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer freader.Close()

	periodSeconds := 1

	switch preriodUnitFlag {
	case "minute":
		periodSeconds = 60
	case "hour":
		periodSeconds = 60 * 60
	case "day":
		periodSeconds = 60 * 60 * 24
	case "m":
		periodSeconds = 60
	case "h":
		periodSeconds = 60 * 60
	case "d":
		periodSeconds = 60 * 60 * 24
	}

	periodSeconds = periodSeconds * int(preriodValueFlag)

	allData := TimeDatas{}

	scanner := bufio.NewScanner(freader)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < len(timeFormatFlag)+2*int(serieCountFlag) {
			continue
		}
		linesplit := strings.Split(line, " ")

		dateStr := strings.Join(linesplit[0:len(linesplit)-int(serieCountFlag)], " ")

		td := TimeData{}
		var err error
		if td.When, err = time.Parse(timeFormatFlag, dateStr); err != nil {
			fmt.Println(err)
			continue
		}

		for _, d := range linesplit[len(linesplit)-int(serieCountFlag) : len(linesplit)] {
			if f, err := strconv.ParseFloat(d, 64); err == nil {
				td.Series = append(td.Series, f)
			} else {
				fmt.Println(err)
				td.Series = append(td.Series, 0.0)
			}
		}

		allData = append(allData, &td)
		//fmt.Printf("%#v\n", td)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading standard input:", err)
	}

	sort.Sort(allData)

	t0 := allData[0].When
	tmax := allData[len(allData)-1].When
	indexMax := 0
	for {
		indexMax = (int(tmax.Sub(t0).Seconds()) / periodSeconds)
		if indexMax > 250 {
			periodSeconds = periodSeconds * 10
			fmt.Printf("Adjust Period: %ds\n", periodSeconds)
		} else {
			break
		}
	}
	fmt.Printf("%d\n", indexMax)
	results := make([]TimeDataAggregated, indexMax+1)

	for _, d := range allData {
		index := int(d.When.Sub(t0).Seconds()) / periodSeconds
		if results[index].Count == 0 {
			results[index].Max = d.Series[0]
			results[index].Min = d.Series[0]
			results[index].Avg = d.Series[0]
			results[index].Count = 1
		} else {
			if d.Series[0] > results[index].Max {
				results[index].Max = d.Series[0]
			}
			if d.Series[0] < results[index].Min {
				results[index].Min = d.Series[0]
			}
			results[index].Avg = (d.Series[0] + results[index].Avg*float64(results[index].Count)) / float64(results[index].Count+1)
			results[index].Count = results[index].Count + 1
		}
	}

	// Build chart
	chart := tm.NewLineChart(tm.Width()-10, tm.Height()-10)
	dataTable := new(tm.DataTable)

	dataTable.AddColumn(fmt.Sprintf("From %s to %s using Block of %d seconds", t0.Format(timeFormatFlag), tmax.Format(timeFormatFlag), periodSeconds))
	if !noMaxFlag {
		dataTable.AddColumn("QMax")
	}
	if !noMinFlag {
		dataTable.AddColumn("QMin")
	}
	if !noAvgFlag {
		dataTable.AddColumn("QAvg")
	}

	cparam := 1
	if !noMaxFlag {
		cparam += 1
	}
	if !noMinFlag {
		cparam += 1
	}
	if !noAvgFlag {
		cparam += 1
	}

	for t := 0; t <= int(indexMax); t++ {
		if results[t].Count != 0 {
			ds := []float64{float64(t)}
			if !noMaxFlag {
				ds = append(ds, results[t].Max)
			}
			if !noMinFlag {
				ds = append(ds, results[t].Min)
			}
			if !noAvgFlag {
				ds = append(ds, results[t].Avg)
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

			//fmt.Printf("Add: %f,%f,%f,%f\n", float64(t), results[t].Max, results[t].Min, results[t].Avg)
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
