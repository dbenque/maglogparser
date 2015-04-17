package magLogParserServer

import (
	"fmt"
	"time"
)

// Serve launch both the HTTP server and the MQ Listening service
func Serve(webPublicIP string, httpPort int, zmqPort int, channelStr chan string, dotNb uint) {
	fmt.Println("Serving")

	var metrics metricArray // Based on benchmark, prefer array approach than heap
	//var metrics metricHeap
	metrics.Init(dotNb, 10)

	go serveZMQ(zmqPort, channelStr, &metrics)
	time.Sleep(time.Second)
	go serveHTTP(webPublicIP, httpPort, &metrics)

	//go ppp(&metrics)

}
