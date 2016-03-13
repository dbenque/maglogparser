package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/dbenque/maglogparser/appdo/api"
	"github.com/dbenque/maglogparser/appdo/parser"
)

func main() {
	doc, err := parser.GetGoqueryDocumentFromFile("appdo.logs")
	if err != nil {
		fmt.Printf("Error while reading file:%v", err)
	}

	records := parser.ExtractData(doc)
	sort.Sort(records)
	parser.CheckDupe(records)
	noDupeRecords := parser.FilterDupe(records)

	filteredStop := noDupeRecords.Filter(&api.Appdoline{Task: "STOP"})
	filteredStart := noDupeRecords.Filter(&api.Appdoline{Task: "START"})
	fmt.Printf("STOP:%d\n", len(filteredStop))
	fmt.Printf("START:%d\n", len(filteredStart))

	for app, recStop := range *filteredStop.SplitPerApps() {
		for _, v := range BlockDetection(recStop, 8*time.Second).Lines {
			fmt.Printf("Blocks %s:%d\n", app, len(v))
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
