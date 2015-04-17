package magLogParserAgent

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/ActiveState/tail"
	zmq "github.com/pebbe/zmq4"
)

type sender interface {
	Send(string, zmq.Flag) (int, error)
}

type channelSender struct {
	channelStr chan string
}

func (mSender *channelSender) Send(s string, f zmq.Flag) (int, error) {
	mSender.channelStr <- s
	return 0, nil
}

// regular expression to apply
var filterRegExp *regexp.Regexp

// RunAgent launch the Agent process
func RunAgent(serverURL string, serverPort int, channelStr chan string, fileToTail string) {

	if channelStr == nil {
		// open the connection with the server
		serverStr := serverURL + ":" + strconv.Itoa(serverPort)
		fmt.Println("Running Agent against server " + serverStr)
		zmqSender, _ := zmq.NewSocket(zmq.PUSH)
		defer zmqSender.Close()
		zmqSender.Connect("tcp://" + serverStr)
		runAgent(zmqSender, fileToTail)
	} else {
		fmt.Println("Running Agent in allInOne mode")
		c := channelSender{channelStr}
		runAgent(&c, fileToTail)
	}

}

func runAgent(aSender sender, fileToTail string) {
	// prepare filter
	compileFilters()

	if len(fileToTail) > 0 {
		// keep watching for the file
		for {
			if t, err := tail.TailFile(fileToTail, tail.Config{Follow: true, MustExist: true}); err != nil {
				fmt.Printf("Error in receive: %v\n", err)
			} else {
				//filterAndSend(zmqSender, t.Lines)
				filterAndBulkSend(aSender, t.Lines)
			}

			// the reopen option is not working in the TailFile config. Let's reopen after 1s in case of file lost
			timer := time.NewTimer(time.Second * 1)
			<-timer.C
		}
	} else { //read from stdin
		fmt.Printf("Reading from STDIN")
		lines := make(chan *tail.Line)
		go func(lines chan *tail.Line) {
			bio := bufio.NewReader(os.Stdin)
			for {
				line, _, err := bio.ReadLine()
				lines <- &(tail.Line{string(line), time.Now(), err})
			}
		}(lines)

		filterAndBulkSend(aSender, lines)
	}

}

func compileFilters() {
	filterRegExp, _ = regexp.Compile("(..../../.. ..:..:..).*Command sequence queue size: ([0-9]+)$")
}

func filterAndSend(aSender sender, lines chan *tail.Line) {
	for line := range lines {
		if matches := filterRegExp.FindStringSubmatch(line.Text); matches != nil {
			aSender.Send(matches[1]+" "+matches[2], 0)
		}
	}
}

func filterAndBulkSend(aSender sender, lines chan *tail.Line) {

	bulkSize := 10
	bulkingTime := time.Second * 3

	// prepare timer for sending data in case bulk is not full (and not empty)
	timer := time.NewTimer(bulkingTime)
	<-timer.C

	noError := true

	// main loop
	for noError {
		i := 0
		bulk := make([]string, bulkSize, bulkSize)
	BULKING:
		for i < bulkSize {
			select {
			// handle incoming lines
			case line, ok := <-lines:
				// exist in case of error with the input lines
				if !ok {
					noError = false
					break BULKING
				}

				// check pattern and bulk if matches
				if matches := filterRegExp.FindStringSubmatch(line.Text); matches != nil {
					bulk[i] = matches[1] + " " + matches[2]
					i++
				}

			// send data even if bulk not full
			case <-timer.C:
				// time to bulk and send
				break BULKING
			}
		}

		// send in case there is something to send
		if i > 0 {
			dataj, _ := json.Marshal(bulk)
			aSender.Send(string(dataj), 0)
		}
	}
}
