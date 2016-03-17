package main

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/dbenque/maglogparser/appdo/api"
	"github.com/dbenque/maglogparser/utils"

	tm "github.com/buger/goterm"
)

type DoubleStop struct {
	ElapseSecond float64
	FirstStop    *api.Appdoline
}

type DoubleStopByElapse []DoubleStop

func (a DoubleStopByElapse) Len() int           { return len(a) }
func (a DoubleStopByElapse) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a DoubleStopByElapse) Less(i, j int) bool { return a[i].ElapseSecond < a[j].ElapseSecond }

// Check node that have been stopped 2 times in a given elapse
func CheckDoubleStop(records api.Appdolines, elapse time.Duration) {

	result := DoubleStopByElapse{}
	for _, r := range records.Filter(api.Appdoline{Task: "stop"}, false) {
		if r.NextForNode != nil && r.NextForNode.Task == "stop" && r.User != "avasup" {
			diff := r.NextForNode.When.Sub(r.When).Seconds()
			if diff < elapse.Seconds() {

				c1 := strings.Split(r.Cmd, " ")
				sort.Strings(c1)
				c2 := strings.Split(r.NextForNode.Cmd, " ")
				sort.Strings(c2)

				if reflect.DeepEqual(c1, c2) {
					result = append(result, DoubleStop{diff, r})
				}
			}
		}
	}

	sort.Sort(result)

	chart := tm.NewLineChart(tm.Width()-10, tm.Height()-10)

	data := new(tm.DataTable)
	data.AddColumn("Double Stop Count")
	data.AddColumn("Elapse (s)")

	for i, rec := range result {
		data.AddRow(float64(i), rec.ElapseSecond)
	}

	tm.Println(chart.Draw(data))
	tm.Flush()

	for _, rec := range result {
		d := rec.ElapseSecond
		doubleStop := rec.FirstStop
		fmt.Printf("Double Stop for %s after %ds:\n\t%s %s (took %ds, rc %d)\t%s\n\t%s %s (took %ds, rc %d)\t%s\n", doubleStop.Host, int(d), doubleStop.User, doubleStop.When.Format(utils.DateFormat), doubleStop.DurationSec, doubleStop.ReturnCode, doubleStop.Cmd, doubleStop.NextForNode.User, doubleStop.NextForNode.When.Format(utils.DateFormat), doubleStop.NextForNode.DurationSec, doubleStop.NextForNode.ReturnCode, doubleStop.NextForNode.Cmd)
	}

	return
}

// Check node that have been stopped 2 times in a given elapse
func CheckDoubleCmd(records api.Appdolines, elapse time.Duration) {
	result := DoubleStopByElapse{}
	for _, r := range records {
		if r.NextForNode != nil && r.NextForNode.Cmd == r.Cmd {
			diff := r.NextForNode.When.Sub(r.When).Seconds()
			if diff < elapse.Seconds() {
				result = append(result, DoubleStop{diff, r})
			}
		}
	}

	sort.Sort(result)

	chart := tm.NewLineChart(tm.Width()-10, tm.Height()-10)

	data := new(tm.DataTable)
	data.AddColumn("Double Stop Count")
	data.AddColumn("Elapse (s)")

	for i, rec := range result {
		data.AddRow(float64(i), rec.ElapseSecond)
	}

	tm.Println(chart.Draw(data))
	tm.Flush()

	for _, rec := range result {
		d := rec.ElapseSecond
		doubleStop := rec.FirstStop
		fmt.Printf("Double Stop for %s after %ds:\n\t%s %s (took %ds, rc %d)\t%s\n\t%s %s (took %ds, rc %d)\t%s\n", doubleStop.Host, int(d), doubleStop.User, doubleStop.When.Format(utils.DateFormat), doubleStop.DurationSec, doubleStop.ReturnCode, doubleStop.Cmd, doubleStop.NextForNode.User, doubleStop.NextForNode.When.Format(utils.DateFormat), doubleStop.NextForNode.DurationSec, doubleStop.NextForNode.ReturnCode, doubleStop.NextForNode.Cmd)
	}
	return
}
