package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/dbenque/maglogparser/appdo/api"
)
import "time"

//http://iaaweb.muc.amadeus.net/cgi-bin/app_do_logviewer?action=set_filters&fromdate=2016+03+11&todate=2016+03+12&app=*&phase=PRD&host=*&user=*&task=*&display_mode=logs
//&http.Client{CheckRedirect: nil}, nil

func BuildURLs(start, end string) map[string]string {
	start2, err := time.Parse("2006/01/02", start)
	if err != nil {
		fmt.Println(err)
		return map[string]string{}
	}
	end2, err := time.Parse("2006/01/02", end)
	if err != nil {
		fmt.Println(err)
		return map[string]string{}
	}

	return buildURLs(start2, end2)

}

func buildURLs(start, end time.Time) map[string]string {

	urls := map[string]string{}
	for start.Before(end) {

		d := start.Format("2006+01+02")
		dd := start.Format("2006-01-02")
		url := fmt.Sprintf("http://iaaweb.muc.amadeus.net/cgi-bin/app_do_logviewer?action=set_filters&fromdate=%s&todate=%s&app=*&phase=PRD&host=*&user=*&task=*&display_mode=logs", d, d)
		urls[dd] = url
		start = start.AddDate(0, 0, 1)
	}

	return urls
}

func FetchAppDoData(start, end string) api.Appdolines {

	var ModePerm os.FileMode = 0777

	os.MkdirAll("data/html", ModePerm)
	os.MkdirAll("data/json", ModePerm)

	urls := BuildURLs(start, end)

	data := api.Appdolines{}

	for d, url := range urls {

		// do we have the html ?
		fHTML := "data/html/" + d + ".html"
		if _, err := os.Stat(fHTML); err != nil {
			if err := fetchHTMLFile(fHTML, url); err != nil {
				fmt.Println(err)
				continue
			}
		}
	}

	dataChan := make(chan api.Appdolines)

	var wgAppend sync.WaitGroup
	wgAppend.Add(1)
	go func() {
		defer wgAppend.Done()
		for a := range dataChan {
			data = append(data, a...)
		}
	}()

	var wg sync.WaitGroup
	for d := range urls {
		wg.Add(1)
		go func(date string) {
			defer wg.Done()
			fHTML := "data/html/" + date + ".html"
			// do we have the json ?
			fJSON := "data/json/" + date + ".json"
			if _, err := os.Stat(fJSON); err != nil {
				doc, err := GetGoqueryDocumentFromFile(fHTML)
				if err != nil {
					fmt.Println(err)
					return
				}
				records := extractData(doc)
				b, err := json.Marshal(records)
				if err != nil {
					fmt.Println(err)
					return
				}
				if err := ioutil.WriteFile(fJSON, b, 0666); err != nil {
					fmt.Println("Can't write json file")
					fmt.Println(err)
					return
				}
				fmt.Printf("Adding %d records from %s\n", len(records), fHTML)
				dataChan <- records
			} else {
				b, err := ioutil.ReadFile(fJSON)
				if err != nil {
					fmt.Println("Can't read json file")
					fmt.Println(err)
					return
				}
				var records api.Appdolines
				if err := json.Unmarshal(b, &records); err != nil {
					fmt.Println(err)
					return
				}
				fmt.Printf("Adding %d records from %s\n", len(records), fJSON)
				dataChan <- records
			}
		}(d)
	}

	wg.Wait()
	close(dataChan)
	wgAppend.Wait()

	fmt.Printf("Count before prepare: %d\n", len(data))

	preparedRecords := prepareData(data)
	fmt.Printf("Count after prepare: %d\n", len(preparedRecords))

	return preparedRecords
}

func fetchHTMLFile(path, url string) error {

	fmt.Printf("Fetching url: %s\n", url)

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(path, contents, 0666); err != nil {
		fmt.Println(err)
	}
	return nil
}
