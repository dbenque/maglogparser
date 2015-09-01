package magLogParserServer

import (
	"encoding/json"
	"fmt"
	"maglogparser/utils"
	"strconv"

	zmq "github.com/pebbe/zmq4"
)

// serveZMQ
func serveZMQ(zmqPort int, channelStr chan string, metrics metricContainer) {

	var zmqListener *zmq.Socket

	if channelStr == nil {
		listenerURL := "*:" + strconv.Itoa(zmqPort)
		fmt.Println("ZeroMQ listening on port: " + listenerURL)
		zmqListener, _ = zmq.NewSocket(zmq.PULL)
		defer zmqListener.Close()
		zmqListener.Bind("tcp://" + listenerURL)
	}

	for {

		var msg string
		if channelStr == nil {
			//  Wait for next request from client
			var err error
			msg, err = zmqListener.Recv(0)
			if err != nil {
				fmt.Printf("Error in receive: %v", err)
				break
			}
		} else {
			msg = <-channelStr
		}

		// unmarshall bulked data
		var bulk []string
		err := json.Unmarshal([]byte(msg), &bulk)
		if err != nil {
			fmt.Println("json unmarshall error:", err)
		}

		// extra data
		for _, data := range bulk {

			dtime, _ := utils.ParseDate4(data[:19])
			//dtime, err := time.Parse(dtFormat, data[:19]) // date time
			if err != nil {
				fmt.Println("time.Parse error:", err)
			}

			value := data[20:]
			intval, _ := strconv.Atoi(value)
			m := metric{dtime, intval}
			metrics.AddMetric(&m)
			//fmt.Println("At ", dtime, " value=", value)

		}

	}

}
