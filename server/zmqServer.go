package magLogParserServer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	zmq "github.com/pebbe/zmq4"
)

const dtFormat = "2006/01/02 15:04:05"

// ParseDate1 parse date using basic Go primitive
func ParseDate1(strdate string) (time.Time, error) {
	return time.Parse(dtFormat, strdate)
}

// ParseDate2 parse date using dedicated function
func ParseDate2(strdate string) (time.Time, error) {
	year, _ := strconv.Atoi(strdate[:4])
	month, _ := strconv.Atoi(strdate[5:7])
	day, _ := strconv.Atoi(strdate[8:10])
	hour, _ := strconv.Atoi(strdate[11:13])
	minute, _ := strconv.Atoi(strdate[14:16])
	second, _ := strconv.Atoi(strdate[17:19])

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC), nil
}

// ParseDate3 parse date using basic Go primitive
func ParseDate3(date []byte) (time.Time, error) {
	year := ((((int(date[0])-'0')*100 + (int(date[1])-'0')*10) + (int(date[2]) - '0')) * 10) + (int(date[3]) - '0')
	month := time.Month(((int(date[5]) - '0') * 10) + (int(date[6]) - '0'))
	day := ((int(date[8]) - '0') * 10) + (int(date[9]) - '0')
	hour := ((int(date[11]) - '0') * 10) + (int(date[12]) - '0')
	minute := ((int(date[14]) - '0') * 10) + (int(date[15]) - '0')
	second := ((int(date[17]) - '0') * 10) + (int(date[18]) - '0')
	return time.Date(year, month, day, hour, minute, second, 0, time.UTC), nil
}

// ParseDate4 parse date optimized
func ParseDate4(date string) (time.Time, error) {
	year := ((((int(date[0])-'0')*100 + (int(date[1])-'0')*10) + (int(date[2]) - '0')) * 10) + (int(date[3]) - '0')
	month := time.Month(((int(date[5]) - '0') * 10) + (int(date[6]) - '0'))
	day := ((int(date[8]) - '0') * 10) + (int(date[9]) - '0')
	hour := ((int(date[11]) - '0') * 10) + (int(date[12]) - '0')
	minute := ((int(date[14]) - '0') * 10) + (int(date[15]) - '0')
	second := ((int(date[17]) - '0') * 10) + (int(date[18]) - '0')
	return time.Date(year, month, day, hour, minute, second, 0, time.UTC), nil
}

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

			dtime, _ := ParseDate4(data[:19])
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
