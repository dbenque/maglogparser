package parser

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dbenque/maglogparser/appdo/api"
)

//URLGetter interface that define the method to retrieve URL
type URLGetter interface {
	Get(url string) (*http.Response, error)
}

//GetGoqueryDocumentFromUrl get document directly from URL
func GetGoqueryDocumentFromUrl(getter URLGetter, url string) (*goquery.Document, error) {

	res, errGet := getter.Get(url)
	if errGet != nil {
		return nil, errGet
	}

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return nil, errGet
	}

	return doc, err
}

//GetGoqueryDocumentFromFile get document directly from URL
func GetGoqueryDocumentFromFile(filePath string) (*goquery.Document, error) {

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		return nil, err
	}
	f.Close()

	return doc, err
}

func ExtractData(doc *goquery.Document) api.Appdolines {
	apps := api.NewApps()
	records := []api.Appdoline{}
	// main array to be browsed
	doc.Find("table").Last().Find("tr").Each(func(i int, s *goquery.Selection) { //row
		vals := []string{}
		s.Find("td").Each(func(j int, sl *goquery.Selection) { // col
			vals = append(vals, sl.Text())
		})

		if len(vals) == 9 {
			var line api.Appdoline
			d, err := time.Parse("Mon Jan 2 15:04:05", vals[1])

			if err == nil {
				line.When = d
				line.App = vals[2]
				line.Host = vals[4]
				line.User = vals[5]
				line.Task = vals[6]
				line.Cmd = vals[8]
				line.Dupe = false
				records = append(records, line)
				apps.Add(line.App)
			}
		}
	})
	api.AddApps(apps)
	return records
}

func CheckDupe(records api.Appdolines) {

	countGoRoutines := 50
	blockOverlap := 200
	blockSize := len(records) / countGoRoutines

	if blockSize < blockOverlap {
		blockSize = 3 * blockOverlap
	}

	start := 0
	fmt.Printf("CheckDupe Analysis (blocksize=%d ; blockOverlap=%d; sliceSize=%d)\n", blockSize, blockOverlap, len(records))
	var wg sync.WaitGroup
	for start < len(records) {
		wg.Add(1)
		go func(startIndex int) {
			fmt.Printf("Dupe Analysis started at index %d\n", startIndex)
			defer wg.Done()
			end := startIndex + blockSize
			if end >= len(records) {
				end = len(records) - 1
			}
			for i, v := range records[startIndex:end] {
				f := i + blockOverlap
				if f > end {
					f = end
				}
				for j := i + 1; j < blockSize; j++ {
					if v.Same(&records[j]) {
						records[j].Dupe = true
						//fmt.Printf("Dupe: \n%v\n%v\n", v, records[j])
					}
				}
			}
		}(start)
		start = start + blockSize - blockOverlap
	}
	wg.Wait()
}

func FilterDupe(records api.Appdolines) api.Appdolines {
	b := records[:0]
	for _, x := range records {
		if !x.Dupe {
			b = append(b, x)
		}
	}
	return b
}
