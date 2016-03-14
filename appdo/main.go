package main

import (
	"fmt"
	"time"

	"github.com/dbenque/maglogparser/appdo/api"
	"github.com/dbenque/maglogparser/appdo/parser"
	"github.com/dbenque/maglogparser/utils"
)

func main() {
	doc, err := parser.GetGoqueryDocumentFromFile("appdo.logs")
	if err != nil {
		fmt.Printf("Error while reading file:%v", err)
	}

	noDupeRecords := parser.ExtractData(doc)

	// filteredStop := noDupeRecords.Filter(&api.Appdoline{Task: "STOP"})
	// filteredStart := noDupeRecords.Filter(&api.Appdoline{Task: "START"})
	// fmt.Printf("STOP:%d\n", len(filteredStop))
	// fmt.Printf("START:%d\n", len(filteredStart))
	// for app, rec := range *noDupeRecords.SplitPerApps() {
	// 	// for _, v := range BlockDetection(rec, 10*time.Second).Lines {
	// 	// 	fmt.Printf("Blocks %s:%d\n", app, len(v))
	// 	// }
	//
	// 	nodes := ChainNodesRecords(rec)
	// 	fmt.Printf("Number of nodes for application %s: %d\n", app, len(nodes))
	// 	for k := range nodes {
	// 		fmt.Printf("%s,", k)
	// 	}
	// 	fmt.Println(" ")
	// }

	APEStop := noDupeRecords.Filter(&api.Appdoline{App: "APE", User: "prdape", Task: "STOP"})
	//APEStop := noDupeRecords.Filter(&api.Appdoline{App: "CML", User: "prdcml", Task: "STOP"})

	for i := range APEStop {
		stop, start, err := GetSequenceDuration(APEStop[i])
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("%s : stop=%f, start=%f\n", APEStop[i].When.Format(utils.DateFormat), stop.Seconds(), start.Seconds())
		}
	}

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
	if stopRecord == nil || stopRecord.Task != "STOP" {
		return 0, 0, fmt.Errorf("Bad input record to compute sequence")
	}

	if stopRecord.NextForNode == nil {
		return 0, 0, fmt.Errorf("Input record not chained")
	}

	if stopRecord.NextForNode.Task != "START" {
		return 0, 0, fmt.Errorf("Input record not chained with classic STOP/START sequence: %s", stopRecord.NextForNode.Task)
	}

	if stopRecord.NextForNode.NextForNode == nil {
		return 0, 0, fmt.Errorf("Input record not chained (2)")
	}

	if !(stopRecord.NextForNode.NextForNode.Task == "RECYCLE" || stopRecord.NextForNode.NextForNode.Task == "RESUME") {
		return 0, 0, fmt.Errorf("Input record not chained with classic STOP/START/(RECYCLE|RESUME) sequence: %s,%s,%s,%s", stopRecord.NextForNode.NextForNode.Task,
			stopRecord.Host, stopRecord.NextForNode.Host, stopRecord.NextForNode.NextForNode.Host)
	}

	return stopRecord.NextForNode.When.Sub(stopRecord.When), stopRecord.NextForNode.NextForNode.When.Sub(stopRecord.NextForNode.When), nil
}
