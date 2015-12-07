package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"maglogparser/locksearch/record"
	rt "maglogparser/locksearch/time"
	"maglogparser/locksearch/window"
	"maglogparser/utils"

	"maglogparser/locksearch/TID"
	"maglogparser/locksearch/cmd"
	"maglogparser/locksearch/lock"
	logging "maglogparser/locksearch/log"
	"maglogparser/locksearch/queue"

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
			chanThreadID <- r
			chanTime <- r
		}
	}

	close(chanThreadID)
	close(chanTime)

	wg.Wait()

	shell.Register("setTime", func(args ...string) (string, error) {
		line := strings.Join(args, " ")
		if strings.Contains(line, "--help") {
			fmt.Println("set the time around which the log exploration should be done.\nTo retrieve logs and on going commands.\nSyntax: setTime " + utils.DateFormat)
			return "", nil
		}

		t, err := time.Parse(utils.DateFormat, line)
		if err != nil {
			return "", err
		}

		if err := window.SetTime(t); err != nil {
			return "", err
		}

		rt.SetTime(t)

		return "", nil
	})

	shell.Register("TID", func(args ...string) (string, error) {
		line := strings.Join(args, " ")
		if strings.Contains(line, "--help") {
			fmt.Println("Show the number of command executed by each thread.\nSyntax: TID")
			return "", nil
		}

		TID.StatTID()
		return "", nil
	})

	shell.Register("log", func(args ...string) (string, error) {
		line := strings.Join(args, " ")
		if strings.Contains(line, "--help") {
			fmt.Println("Display the log around the time that was set (setTime).\nSyntax: log TID [nbLine=5]")
			return "", nil
		}

		if len(args) < 1 {
			fmt.Println("Syntax error, missing arguments")
			fmt.Println("Display the log around the time that was set (setTime).\nSyntax: log TID [nbLine=5]")
			return "", nil
		}

		count := 0
		if len(args) == 2 {
			if c, err := strconv.Atoi(args[1]); err == nil {
				count = c
			}
		}

		logging.GetLog(args[0], count)
		return "", nil
	})

	shell.Register("queue", func(args ...string) (string, error) {
		line := strings.Join(args, " ")
		if strings.Contains(line, "--help") {
			fmt.Println("Build statistic for queues.\nSyntax: queue")
			return "", nil
		}

		queue.StatQueue()
		return "", nil
	})

	shell.Register("cmd", func(args ...string) (string, error) {
		line := strings.Join(args, " ")
		if strings.Contains(line, "--help") {
			fmt.Println("Show the commands statistics.\nSyntax: cmd")
			fmt.Println("Show the commands for a given thread in the time window.\nSyntax: cmd TID")
			return "", nil
		}

		if len(args) == 1 {
			if _, err := strconv.Atoi(args[0]); err == nil {
				cmd.StatCmdTID(args[0])
				return "", nil
			}
		}

		cmd.StatCmd()
		return "", nil
	})

	shell.Register("lock", func(args ...string) (string, error) {
		line := strings.Join(args, " ")
		if strings.Contains(line, "--help") {
			fmt.Println("Search for each thread the max hang time.\nSyntax: lock [cmd] [setWindow]\n\tcmd: Only display the theard that execute commands\n\tsetWindow: set the time window to the range of time returned by this lock exploration.")
			return "", nil
		}

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

		lock.StatLock(onlyCmd, setWindow)

		if setWindow {
			updatePromt()
		}

		return "", nil
	})

	shell.Register("window", func(args ...string) (string, error) {
		line := strings.Join(args, " ")
		if strings.Contains(line, "--help") {
			fmt.Println("Display the time window of all the logs and the enclosed one used for the statistics.\nSyntax: window [reset]\n\treset: set the active window to the maximum (all logs)")
			return "", nil
		}

		if len(args) > 0 && args[0] == "reset" {
			window.Reset()
			updatePromt()
		}

		return window.GetWindow().Print(), nil
	})

	shell.Register("start", func(args ...string) (string, error) {
		line := strings.Join(args, " ")
		if strings.Contains(line, "--help") {
			fmt.Println("Set the lower bound of the active time window.\nSyntax: start " + utils.DateFormat)
			return "", nil
		}

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
		line := strings.Join(args, " ")
		if strings.Contains(line, "--help") {
			fmt.Println("Set the upper bound of the active time window.\nSyntax: end " + utils.DateFormat)
			return "", nil
		}

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
