package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/dbenque/maglogparser/appdo/api"
	"github.com/dbenque/maglogparser/appdo/parser"
	"github.com/dbenque/maglogparser/utils"

	tm "github.com/buger/goterm"
)

func main() {

	fmt.Println("Reading Data")
	records := parser.FetchAppDoData("2015/06/01", "2016/03/15")
	fmt.Println("Fetch complete")

	// APEStop := records.Filter(&api.Appdoline{App: "ape", User: "prdape", Task: "stop"})
	//
	// for i := range APEStop {
	// 	stop, start, err := GetSequenceDuration(APEStop[i])
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	} else {
	// 		fmt.Printf("%s : stop=%f, start=%f\n", APEStop[i].When.Format(utils.DateFormat), stop.Seconds(), start.Seconds())
	// 	}
	// }
	//CheckDoubleStop(records, 10*time.Minute)
	// for _, doubleStopStruct := range  {
	// 	d := doubleStopStruct.ElapseSecond
	// 	doubleStop := doubleStopStruct.FirstStop
	// 	fmt.Printf("Double Stop for %s after %ds:\n\t%s %s (took %ds, rc %d)\t%s\n\t%s %s (took %ds, rc %d)\t%s\n", doubleStop.Host, d, doubleStop.User, doubleStop.When.Format(utils.DateFormat), doubleStop.DurationSec, doubleStop.ReturnCode, doubleStop.Cmd, doubleStop.NextForNode.User, doubleStop.NextForNode.When.Format(utils.DateFormat), doubleStop.NextForNode.DurationSec, doubleStop.NextForNode.ReturnCode, doubleStop.NextForNode.Cmd)
	// }

	//CheckDoubleCmd(records, 3*time.Second)

	for _, app := range []string{"adv", "rfd", "ess", "tds", "aml", "loy"} {
		// 	// 	fmt.Printf("\n-------------------- %s -----------------------\n", app)
		// 	//
		// 	// 	DurationTrendGraph(records.Filter(api.Appdoline{App: app, Task: "stop", ReturnCode: 0}, false).Filter(api.Appdoline{User: "avasup"}, true), "stop time "+app)
		// 	// 	DurationTrendGraph(records.Filter(api.Appdoline{App: app, Task: "start", ReturnCode: 0}, false).Filter(api.Appdoline{User: "avasup"}, true), "start time "+app)
		DurationTrendGraph(records.Filter(api.Appdoline{App: app, Task: "start", ReturnCode: 0}, false).Filter(api.Appdoline{User: "avasup"}, true), app)
	}

	//DisplayAll(records.Filter(api.Appdoline{App: "tds", Task: "start", Host: "obeap161", ReturnCode: 0}, false).Filter(api.Appdoline{User: "avasup"}, true))
	//DisplayAll(records.Filter(api.Appdoline{App: "tds", Task: "start", Host: "obeap345", ReturnCode: 0}, false).Filter(api.Appdoline{User: "avasup"}, true))
	//isplayAll(records.Filter(api.Appdoline{User: "pjzenoz"}, false))
	//DisplayAll(records)

}

type Blocks struct {
	Elapse time.Duration
	Lines  []api.Appdolines
}

func BlockDetection(records api.Appdolines, elapse time.Duration) *Blocks {
	blocks := Blocks{Elapse: elapse, Lines: []api.Appdolines{}}
	blockStart := 0
	for i := 1; i < len(records); i++ {
		if records[i].When.Sub(records[i-1].When) > elapse {
			blocks.Lines = append(blocks.Lines, records[blockStart:i])
			blockStart = i
		}
	}
	blocks.Lines = append(blocks.Lines, records[blockStart:len(records)])
	return &blocks
}

func GetSequenceDuration(stopRecord *api.Appdoline) (stopTime, startTime time.Duration, err error) {
	if stopRecord == nil || stopRecord.Task != "stop" {
		return 0, 0, fmt.Errorf("Bad input record to compute sequence")
	}

	if stopRecord.NextForNode == nil {
		return 0, 0, fmt.Errorf("Input record not chained after STOP")
	}

	if stopRecord.NextForNode.Task != "start" {
		return 0, 0, fmt.Errorf("Input record not chained with classic STOP/START sequence: %s", stopRecord.NextForNode.Task)
	}

	if stopRecord.NextForNode.NextForNode == nil {
		return 0, 0, fmt.Errorf("Input record not chained after START, %s, %s, %s", stopRecord.NextForNode.App, stopRecord.NextForNode.Host, stopRecord.NextForNode.When.Format(utils.DateFormat))
	}

	if !(stopRecord.NextForNode.NextForNode.Task == "cycle" || stopRecord.NextForNode.NextForNode.Task == "resume") {
		return 0, 0, fmt.Errorf("Input record not chained with classic STOP/START/(RECYCLE|RESUME) sequence: %s,%s,%s,%s", stopRecord.NextForNode.NextForNode.Task,
			stopRecord.Host, stopRecord.NextForNode.Host, stopRecord.NextForNode.NextForNode.Host)
	}

	return stopRecord.NextForNode.When.Sub(stopRecord.When), stopRecord.NextForNode.NextForNode.When.Sub(stopRecord.NextForNode.When), nil
}

func DisplayAll(records api.Appdolines) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 20, 0, 2, ' ', tabwriter.AlignRight)

	fmt.Fprintln(w, "Time\tApp\tNode\tUser\tTask\tCommand\tRC\tDuration\t")

	for _, r := range records {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%d\t%d\t\n", r.When.Format(utils.DateFormat), r.App, r.Host, r.User, r.Task, r.Cmd, r.ReturnCode, r.DurationSec)
	}

	fmt.Fprintln(w)
	w.Flush()
}

func DurationTrendGraph(records api.Appdolines, dataTitle string) {

	chart := tm.NewLineChart(tm.Width()-10, tm.Height()-10)

	data := new(tm.DataTable)
	if len(records) > 0 {
		data.AddColumn(records[0].When.Format("2006/01/02") + " --> " + records[len(records)-1].When.Format("2006/01/02"))
	} else {
		data.AddColumn("")
	}

	data.AddColumn(dataTitle)

	for i, rec := range records {
		if rec.DurationSec > 6000 {
			continue
		}
		data.AddRow(float64(i), float64(rec.DurationSec))
	}

	tm.Println(chart.Draw(data))
	tm.Flush()

}
