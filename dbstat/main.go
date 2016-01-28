package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"text/tabwriter"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	if len(os.Args) != 2 {
		log.Fatal("Missing parameter.\nUsage: " + os.Args[0] + " dbfile\n")
		return
	}

	db, err := sql.Open("sqlite3", os.Args[1])
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	{
		rows, err := db.Query("select count(*) from var")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var c int

			rows.Scan(&c)
			fmt.Printf("All variable count: %d\n\n", c)
		}
	}

	sliceAppStat := []appStat{}

	printCount(db)
	fmt.Println("##############################################")
	fmt.Println("############# Per Application Analysis #######")
	fmt.Println("##############################################")
	fmt.Println("")
	{
		rows, err := db.Query("select name from application")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		type resultPerApp struct {
			appName     string
			duplicate   int
			bytes       int
			activateVar int
			buffer      *bytes.Buffer
		}

		chanResult := make(chan resultPerApp)
		done := make(chan bool)
		wg := sync.WaitGroup{}
		for rows.Next() {
			var appName string

			rows.Scan(&appName)
			wg.Add(1)
			go func(appN string) {
				c, b, a, buf := printStatOnValues(collectAppVar(db, appN))
				chanResult <- resultPerApp{appN, c, b, a, buf}
				wg.Done()
			}(appName)

		}

		go func() {
			for {
				select {
				case r := <-chanResult:
					fmt.Printf("\n\n##################################### %s\n\n", r.appName)

					fmt.Printf("%s", r.buffer.String())
					sliceAppStat = append(sliceAppStat, appStat{r.appName, r.duplicate, r.bytes, r.activateVar})

				case <-done:
					return
				}
			}
		}()

		wg.Wait()
		done <- true
	}
	// appName := "ETK"
	// c, b := printStatOnValues(collectAppVar(db, appName))
	// sliceAppStat = append(sliceAppStat, appStat{appName, c, b})

	fmt.Printf("\n\n")
	fmt.Println("##############################################")
	fmt.Println("############# Statistics############## #######")
	fmt.Println("##############################################")

	sort.Sort((byCount)(sliceAppStat))

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 20, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "Application\tDuplication\tPossible Saved Bytes\tPossible obsolete Activation\t")

	for _, stat := range sliceAppStat {
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t\n", stat.name, stat.repetions, stat.bytes, stat.activation)
	}

	fmt.Fprintln(w)
	w.Flush()

	fmt.Println()
	fmt.Println()

}

type appStat struct {
	name       string
	repetions  int
	bytes      int
	activation int
}

type byCount []appStat

func (a byCount) Len() int           { return len(a) }
func (a byCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byCount) Less(i, j int) bool { return a[i].bytes < a[j].bytes }
