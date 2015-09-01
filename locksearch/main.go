package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"maglogparser/locksearch/record"
	"maglogparser/locksearch/window"
	"maglogparser/utils"

	"github.com/abiosoft/ishell"
)

var flagvar int
var shell *ishell.Shell

func init() {
	flag.IntVar(&flagvar, "flagname", 1234, "help message for flagname")
	shell = ishell.NewShell()
}

func main() {

	args := os.Args[1:]
	var records []*record.Record

	chanThreadID := make(chan *record.Record)
	chanTime := make(chan *record.Record)

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		record.ScanThreadID(chanThreadID)
	}()
	go func() {
		defer wg.Done()
		window.ScanTime(chanTime)
	}()

	for s := range readLine(args[0]) {
		r := record.NewRecord(s)
		if r != nil {
			records = append(records, r)
			chanThreadID <- r
			chanTime <- r
		}
	}

	close(chanThreadID)
	close(chanTime)

	wg.Wait()

	shell.Register("setTime", func(args ...string) (string, error) {
		t, err := time.Parse(utils.DateFormat, strings.Join(args, " "))

		if err != nil {
			return "", err
		}

		if err := window.SetTime(t); err != nil {
			return "", err
		}

		setTime(t)

		return "", nil
	})

	shell.Register("TID", func(args ...string) (string, error) {
		statTID()
		return "", nil
	})

	shell.Register("cmd", func(args ...string) (string, error) {
		statCmd()
		return "", nil
	})

	shell.Register("lock", func(args ...string) (string, error) {
		onlyCmd := false
		setWindow := false
		for _, a := range args {
			if a == "cmd" {
				onlyCmd = true
			}
			if a == "setWindow" {
				setWindow = true
			}
		}

		statLock(onlyCmd, setWindow)

		if setWindow {
			updatePromt()
		}

		return "", nil
	})

	shell.Register("window", func(args ...string) (string, error) {

		if len(args) > 0 && args[0] == "reset" {
			window.Reset()
			updatePromt()
		}

		return window.GetWindow().Print(), nil
	})

	shell.Register("start", func(args ...string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("start take 1 argument that must be a timestamp")
		}

		start, err := time.Parse(utils.DateFormat, strings.Join(args, " "))

		if err != nil {
			return "", err
		}
		if err = window.SetStart(start); err != nil {
			return "", err
		}
		updatePromt()

		return "", nil
	})

	shell.Register("end", func(args ...string) (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("start take 1 argument that must be a timestamp")
		}

		start, err := time.Parse(utils.DateFormat, strings.Join(args, " "))

		if err != nil {
			return "", err
		}
		if err = window.SetEnd(start); err != nil {
			return "", err
		}
		updatePromt()

		return "", nil
	})

	updatePromt()

	shell.Start()

}

func updatePromt() {
	shell.SetPrompt("\n" + window.GetWindow().PrintCurrent() + " >> ")
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
