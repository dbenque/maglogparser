package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
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

	for _, files := range args {
		for s := range readLine(files) {
			r := record.NewRecord(s)
			if r != nil {
				chanThreadID <- r
				chanTime <- r
			}
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
			fmt.Println("Display nbLine log lines around the time that was set (setTime) for a given TID.\nSyntax: log TID [nbLine=5]")
			fmt.Println("Display the log of the command for the time that was set (setTime) for a given TID.\nSyntax: log TID cmd")
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
			} else if args[1] == "cmd" {
				count = -1
			}
		}

		logging.GetLog(args[0], count)
		return "", nil
	})

	shell.Register("queue", func(args ...string) (string, error) {
		line := strings.Join(args, " ")
		if strings.Contains(line, "--help") {
			fmt.Println("Build statistic for queues (readonly pool or write).\nSyntax: queue [r|w]")
			return "", nil
		}

		queue.StatQueue(len(args) > 0 && args[0] == "w")
		return "", nil
	})

	shell.Register("cmd", func(args ...string) (string, error) {
		line := strings.Join(args, " ")
		if strings.Contains(line, "--help") {
			fmt.Println("Show the commands statistics.\nSyntax: cmd")
			fmt.Println("Show the commands for a given thread in the time window.\nSyntax: cmd 'TID'")
			fmt.Println("Show the response time distribution in the time window.\nSyntax: cmd d 'CmdName'")
			return "", nil
		}

		if len(args) == 1 {
			if _, err := strconv.Atoi(args[0]); err == nil {
				cmd.StatCmdTID(args[0])
				return "", nil
			}
		}

		if len(args) == 2 {
			if args[0] == "d" {
				cmd.StatCmdDistribution(args[1])
				return "", nil
			}
		}

		cmd.StatCmdAll()

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

	scanner, err := scannerFromFile(inFile)
	if err != nil {
		log.Fatalf("Scanner type error : %v", err)
		return nil
	}

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

func scannerFromFile(reader io.Reader) (*bufio.Scanner, error) {

	var scanner *bufio.Scanner
	//create a bufio.Reader so we can 'peek' at the first few bytes
	bReader := bufio.NewReader(reader)

	testBytes, err := bReader.Peek(64) //read a few bytes without consuming
	if err != nil {
		return nil, err
	}
	//Detect if the content is gzipped
	contentType := http.DetectContentType(testBytes)

	fmt.Printf("Content Type:%s\n", contentType)

	//If we detect gzip, then make a gzip reader, then wrap it in a scanner
	if strings.Contains(contentType, "x-gzip") {
		gzipReader, err := gzip.NewReader(bReader)
		if err != nil {
			return nil, err
		}

		scanner = bufio.NewScanner(gzipReader)

	} else {
		//Not gzipped, just make a scanner based on the reader
		scanner = bufio.NewScanner(bReader)
	}

	return scanner, nil
}
