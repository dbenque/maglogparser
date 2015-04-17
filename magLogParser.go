package main

import (
	"flag"
	"magLogParser/client"
	"magLogParser/server"
)

var moduleFlag = flag.String("module", "wqueue", "module to be used for log parsing: [wqueue]")
var serverModeFlag = flag.Bool("serverMode", false, "specify if the instance should run as server, ie parser. If not it is acting as client, collecting data, and serving")
var webserverURLFlag = flag.String("IP", "", "Public IP where server can be reached")
var zmqPortFlag = flag.Int("dataPort", 41000, "Port used to recieve/send data")
var httpPortFlag = flag.Int("httpPort", 8080, "Port used to by webserver")
var fileToTailFlag = flag.String("tail", "", "File to tail when running client mode. If not set read is done on stdin")
var allInOneFlag = flag.Bool("allInOne", false, "Run both the Client  and the server on the same machine")
var pointFlag = flag.Uint("dot", 100, "Number of dot saved")

func init() {
	flag.StringVar(moduleFlag, "m", "wqueue", "shortcut for -module")
}

func main() {
	flag.Parse()
	//fmt.Printf("Selected Module: %s\n", *moduleFlag)

	var comChan chan string

	if *allInOneFlag {
		comChan = make(chan string)
	}

	if *serverModeFlag || *allInOneFlag {
		go magLogParserServer.Serve(*webserverURLFlag, *httpPortFlag, *zmqPortFlag, comChan, *pointFlag)
	}

	if !(*serverModeFlag) || *allInOneFlag {
		go magLogParserAgent.RunAgent(*webserverURLFlag, *zmqPortFlag, comChan, *fileToTailFlag)
	}

	done := make(chan bool)
	<-done
}
