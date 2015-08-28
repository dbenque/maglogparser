package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

const timeFormat = "2006/01/02 15:04:05.000000"

type record struct {
	time             time.Time
	threadId         string
	node             string
	pid              string
	raw              string
	index            int
	nextByThread     *record
	previousByThread *record
}

func NewRecord(raw string) *record {
	if t, err := time.Parse(timeFormat, raw[0:len(timeFormat)]); err == nil {
		r := record{time: t, raw: raw}
		tokens := strings.Split(r.raw, " ")
		r.node = tokens[2]
		r.pid = tokens[4]
		tokens = strings.Split(strings.Split(r.raw, "<")[1], " ")
		r.threadId = tokens[1][4 : len(tokens[1])-1]

		if r.threadId[0] == 'f' {
			log.Fatalln(r.raw)
		}

		return &r
	}
	return nil
}

var RecordByThread map[string][]*record

func init() {
	RecordByThread = make(map[string][]*record)
}

func main() {

	args := os.Args[1:]
	var records []*record

	index := 0

	chanThreadID := make(chan *record)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		scanThreadID(chanThreadID)
		statThreadID()
	}()

	for s := range readLine(args[0]) {
		r := NewRecord(s)
		if r != nil {
			r.index = index
			records = append(records, r)
			chanThreadID <- r
			index++
		}
	}

	close(chanThreadID)

	wg.Wait()

	fmt.Println("Done")
}

func scanThreadID(c <-chan *record) {

	for r := range c {
		m, ok := RecordByThread[r.threadId]
		if !ok {
			m = make([]*record, 0, 0)
		}

		if len(m) > 0 {
			m[len(m)-1].nextByThread = r
			r.previousByThread = m[len(m)-1]
		}

		m = append(m, r)
		RecordByThread[r.threadId] = m
	}
}

func statThreadID() {
	var wg sync.WaitGroup
	for id, sl := range RecordByThread {
		wg.Add(1)
		go func(tid string, rs []*record) {
			defer wg.Done()
			max := time.Duration(0)
			var rmax *record
			for _, r := range rs {

				if r.nextByThread != nil {
					d := r.nextByThread.time.Sub(r.time)
					if d > max {
						max = d
						rmax = r
					}
				}
			}

			if rmax != nil {
				rcmd := getCurrentCommand(rmax)
				cmd := "nil"
				cmdTime := "nil"

				if rcmd != nil {
					cmd = strings.Split(strings.Split(rcmd.raw, "> ")[1], " ")[2]

					if strings.Contains(cmd, "kSEIAdminCmdGeneric") {
						gen := strings.Split(strings.Split(rcmd.raw, "[")[1], "]")[0]
						cmd = cmd + "[" + gen + "]"
					}

					cmdTime = rcmd.time.Format(timeFormat)
				}

				rcmdComplete := getCommandCompletion(rmax)
				cmdCompleteTime := "nil"
				if rcmdComplete != nil {
					cmdCompleteTime = rcmdComplete.time.Format(timeFormat)
				}

				duration := ""
				if rcmdComplete != nil && rcmd != nil {
					duration = fmt.Sprintf("%f", rcmdComplete.time.Sub(rcmd.time).Seconds())
				}

				fmt.Printf("%s --> TID%s : wait %f till %s for %s since %s to finish at %s duration %s\n", rmax.time.Format(timeFormat), tid, max.Seconds(), rmax.nextByThread.time.Format(timeFormat), cmd, cmdTime, cmdCompleteTime, duration)
			}

		}(id, sl)
	}
	wg.Wait()
}

func getCurrentCommand(r *record) *record {

	for !strings.Contains(r.raw, "> Processing Command#") && r.previousByThread != nil {
		r = r.previousByThread
	}

	if strings.Contains(r.raw, "> Processing Command#") {
		return r
	}

	return nil
}

func getCommandCompletion(r *record) *record {

	for !strings.Contains(r.raw, "> End of processing for Command#") && r.nextByThread != nil {
		r = r.nextByThread
	}

	if strings.Contains(r.raw, "> End of processing for Command#") {
		return r
	}

	return nil
}

func readLine(path string) <-chan string {

	inFile, err := os.Open(path)
	if err != nil {
		log.Fatalf("os error: %v", err)
		return nil
	}

	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)

	c := make(chan string)

	go func() {
		fmt.Printf("Scanning file: %s\n", path)
		for scanner.Scan() {
			c <- scanner.Text()
		}
		close(c)
		inFile.Close()
	}()

	return c
}
