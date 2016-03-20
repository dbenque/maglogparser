package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dbenque/maglogparser/timeChart/timeChart"
)

func init() {
	InitFlags()
}

func main() {

	if ParseFlags() != nil {
		return
	}

	fmt.Printf("Time format: %s\nPeriod: %d%s\n", timeFormatFlag, periodValueFlag, periodUnitFlag)

	freader, err := os.Open(fileFlag)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer freader.Close()

	periodSeconds := 1

	switch periodUnitFlag {
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

	periodSeconds = periodSeconds * int(periodValueFlag)
	tc := timeChart.MakeChart(periodSeconds, timeFormatFlag, !noMaxFlag, !noMinFlag, !noAvgFlag)

	scanner := bufio.NewScanner(freader)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < len(timeFormatFlag)+2*int(serieCountFlag) {
			continue
		}
		linesplit := strings.Split(line, " ")

		dateStr := strings.Join(linesplit[0:len(linesplit)-int(serieCountFlag)], " ")

		td := timeChart.TimeData{}
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
		tc.Add(&td)
	}

	tc.DataCompleted()

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading standard input:", err)
	}

	tc.Draw()

}
