package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

type statVar struct {
	AppName      string
	AppOnlyLevel int
	AppAsLevel   int
	AsBeLevel    int
	BeOnlyLevel  int
	TotalApp     int
}

type byAppName []statVar

func (a byAppName) Len() int           { return len(a) }
func (a byAppName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byAppName) Less(i, j int) bool { return strings.Compare(a[i].AppName, a[j].AppName) < 0 }

type byTotal []statVar

func (a byTotal) Len() int           { return len(a) }
func (a byTotal) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byTotal) Less(i, j int) bool { return a[i].TotalApp < a[j].TotalApp }

var mapStat = map[string]statVar{}

func getStatForApp(app string) statVar {

	if r, ok := mapStat[app]; ok {
		return r
	}
	return statVar{}

}

func setStatForApp(s statVar, app string) {
	mapStat[app] = s
}

func printCount(db *sql.DB) {
	//APPOnly
	{
		rows, err := db.Query("select a.name,count(v.id) from var as v ,application as a where v.appli>0 and v.backendtype=0 and v.appsrv=0 and a.id=v.appli group by a.name")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var appName string
			var c int
			rows.Scan(&appName, &c)
			s := getStatForApp(appName)
			s.AppOnlyLevel = c
			setStatForApp(s, appName)
		}
	}
	// AppAs
	{
		rows, err := db.Query("select a.name,count(v.id) from var as v ,application as a where v.appli>0 and v.backendtype=0 and v.appsrv>0 and a.id=v.appli group by a.name")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var appName string
			var c int
			rows.Scan(&appName, &c)
			s := getStatForApp(appName)
			s.AppAsLevel = c
			setStatForApp(s, appName)
		}
	}

	//AsBe
	{
		rows, err := db.Query("select a.name,count(v.id) from var as v ,application as a, applicationserver as appsrv where v.appli=0 and v.backendtype>0 and v.appsrv=appsrv.id and a.id=appsrv.appli group by a.name")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var appName string
			var c int
			rows.Scan(&appName, &c)
			s := getStatForApp(appName)
			s.AsBeLevel = c
			setStatForApp(s, appName)
		}
	}

	{
		rows, err := db.Query("select name,count(id) from (select a.name,v.id from var as v ,application as a, applicationserver as appsrv , asbetype where v.appli=0 and v.appsrv=0 and v.backendtype=asbetype.betype and a.id=appsrv.appli and appsrv.id=asbetype.appsrv group by a.name,v.id) group by name")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var appName string
			var c int
			rows.Scan(&appName, &c)
			s := getStatForApp(appName)
			s.BeOnlyLevel = c
			setStatForApp(s, appName)
		}
	}

	total := statVar{}

	total.AppName = "Total"
	sliceOfStat := []statVar{}

	for appName, stat := range mapStat {
		stat.AppName = appName
		stat.TotalApp = stat.AppOnlyLevel + stat.AppAsLevel + stat.AsBeLevel + stat.BeOnlyLevel
		total.AppOnlyLevel += stat.AppOnlyLevel
		total.AppAsLevel += stat.AppAsLevel
		total.AsBeLevel += stat.AsBeLevel
		total.BeOnlyLevel += stat.BeOnlyLevel
		total.TotalApp += stat.TotalApp

		sliceOfStat = append(sliceOfStat, stat)
	}

	fmt.Println("############# Count Total ####################")
	printTable([]statVar{total})

	fmt.Println("############# Count By AppName ###############")
	sort.Sort((byAppName)(sliceOfStat))
	printTable(sliceOfStat)
	fmt.Println("############# Count By Total #################")
	sort.Sort((byTotal)(sliceOfStat))
	printTable(sliceOfStat)

}

func printTable(s []statVar) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 20, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "Application\tAppOnly\tAppAs\tAsBe\tBeOnly\tTotal\t")

	for _, stat := range s {
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%d\t\n", stat.AppName, stat.AppOnlyLevel, stat.AppAsLevel, stat.AsBeLevel, stat.BeOnlyLevel, stat.TotalApp)
	}

	fmt.Fprintln(w)
	w.Flush()

	fmt.Println()
	fmt.Println()

}
