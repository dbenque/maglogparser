package magLogParserServer

import (
	"encoding/json"
	"errors"

	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func serveError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, "Internal Server Error")
}

func handleTestOk(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Nice Test"))
}
func handleTestKo(w http.ResponseWriter, r *http.Request) {
	serveError(w, errors.New("MyError"))
}

type serie [][2]int64
type series struct {
	Max     serie
	Min     serie
	Average serie
}

func newSeries(count int) series {
	return series{make(serie, count, count), make(serie, count, count), make(serie, count, count)}
}

var latestPoolTime time.Time
var latestSerie series

func handleData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if latestPoolTime.IsZero() || time.Now().Sub(latestPoolTime).Seconds() > 1 {
		latestPoolTime = time.Now()
		metrics := theHTTPMetricContainer.GetAllMetrics()
		theSeries := newSeries(len(metrics))
		j := 0
		for _, e := range metrics {
			if !e.Timestamp.IsZero() {

				tms := 1000 * e.Timestamp.Unix()
				theSeries.Max[j][0] = tms
				theSeries.Max[j][1] = int64(e.Max)
				theSeries.Min[j][0] = tms
				theSeries.Min[j][1] = int64(e.Min)
				theSeries.Average[j][0] = tms
				if e.Count != 0 {
					theSeries.Average[j][1] = e.Sum / int64(e.Count)
				}
				j++
			}
		}
		latestSerie = series{theSeries.Max[:j], theSeries.Min[:j], theSeries.Average[:j]}
	} else {
	}

	if dataj, err := json.Marshal(latestSerie); err == nil {
		w.Write(dataj)
		w.WriteHeader(http.StatusOK)

	} else {
		fmt.Println("error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Error While Serving Data")
	}

}

// func handleData2(w http.ResponseWriter, r *http.Request) {
//
// 	type jsonMetricSummary struct {
// 		T       int64
// 		Max     int
// 		Min     int
// 		Average int
// 	}
//
// 	w.Header().Set("Content-Type", "application/json")
//
// 	metrics := theHTTPMetricContainer.GetLatestMetrics()
// 	summaries := make([]jsonMetricSummary, len(metrics))
// 	for i, e := range metrics {
// 		summaries[i].T = e.Timestamp.Unix()
// 		summaries[i].Max = e.Max
// 		summaries[i].Min = e.Min
// 		if e.Count != 0 {
// 			summaries[i].Average = int(e.Sum / int64(e.Count))
// 		}
// 	}
//
// 	if dataj, err := json.Marshal(summaries); err == nil {
// 		fmt.Println("WTF:", string(dataj))
// 		w.Write(dataj)
// 		w.WriteHeader(http.StatusOK)
// 		fmt.Println("Serving Data")
// 	} else {
// 		fmt.Println("error:", err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		fmt.Println("Error While Serving Data")
// 	}
//
// }

func handlePage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := GetPageTemplate()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Error While Using Template")
		return
	}
	tmpl.ExecuteTemplate(w, "PAGE", theWebPublicIP+":"+theWebPort)

}

var theHTTPMetricContainer metricContainer
var theWebPublicIP string
var theWebPort string

// serveHTTP launch HTTP Server for magLogParser
func serveHTTP(webPublicIP string, portWeb int, metrics metricContainer) {

	theHTTPMetricContainer = metrics
	theWebPublicIP = webPublicIP
	theWebPort = strconv.Itoa(portWeb)

	r := mux.NewRouter()
	r.HandleFunc("/testOk", handleTestOk)
	r.HandleFunc("/data", handleData)
	r.HandleFunc("/page", handlePage)
	r.HandleFunc("/", handleData)
	//	r.HandleFunc("/page", handlePage)
	http.Handle("/", r)

	fs := http.FileServer(http.Dir("/home/david/dev/go/src/magLogParser/server/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	portstr := strconv.Itoa(portWeb)
	err := http.ListenAndServe(":"+portstr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
