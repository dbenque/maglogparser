package parser

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
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

func extractData(doc *goquery.Document) api.Appdolines {
	apps := api.NewApps()
	records := []*api.Appdoline{}
	// main array to be browsed
	doc.Find("table").Last().Find("tr").Each(func(i int, s *goquery.Selection) { //row
		vals := []string{}
		linkid := ""
		s.Find("td").Each(func(j int, sl *goquery.Selection) { // col
			vals = append(vals, sl.Text())
			if j == 0 {
				sl.Find("input").Each(func(k int, si *goquery.Selection) {
					link, _ := si.Attr("value")
					if len(link) > 0 {
						linkid = strings.Split(link, " ")[0]
					}
				})
			}
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
				line.Link = linkid
				records = append(records, &line)
				apps.Add(line.App)
			}
		}
	})
	api.AddApps(apps)
	return records
}

func CheckDupe(records api.Appdolines) {
	for _, r := range records {
		if r.NextForNode != nil && r.Link == r.NextForNode.Link {
			r.NextForNode.Dupe = true
			r.NextForNode = r.NextForNode.NextForNode
		}
	}
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

func ExtractData(doc *goquery.Document) api.Appdolines {
	records := extractData(doc)
	return prepareData(records)
}

func prepareData(records api.Appdolines) api.Appdolines {
	sort.Sort(records)

	var wg sync.WaitGroup
	applications := api.NewApps()
	for app, rec := range *records.SplitPerApps() {
		applications.Add(app)
		wg.Add(1)
		go func(appli string, appRec api.Appdolines) {
			defer wg.Done()
			nodes := ChainNodesRecords(appRec)
			fmt.Printf("Number of nodes for application %s: %d\n", appli, len(nodes))
			// for k := range nodes {
			// 	fmt.Printf("%s,", k)
			// }
			CheckDupe(appRec)
		}(app, rec)
	}
	wg.Wait()
	api.AddApps(applications)
	fmt.Printf("Applications: %v\n", api.GetApps())
	return FilterDupe(records)
}

// ChainNodesRecords set the NextForNode pointer.
// The given Appdolines must be time sorted.
func ChainNodesRecords(records api.Appdolines) map[string]*api.Appdoline {

	nodes := map[string]*api.Appdoline{}
	nodeStart := map[string]*api.Appdoline{}

	for i, r := range records {
		if rec, ok := nodes[r.Host]; ok {
			rec.NextForNode = records[i]
			nodes[r.Host] = records[i]
		} else {
			nodeStart[r.Host] = records[i]
			nodes[r.Host] = records[i]
		}
	}

	return nodeStart
}
