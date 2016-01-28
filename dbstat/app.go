package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"text/tabwriter"
)

type varObj struct {
	id    int
	name  string
	peak  int
	value string
}

const (
	AppOnlyLevel = 3
	AppAsLevel   = 2
	BeOnlyLevel  = 1
	AsBeLevel    = 0
)

type vars struct {
	Levels [4][]*varObj
}

func (v *vars) count() int {
	c := 0
	for _, a := range v.Levels {
		c += len(a)
	}
	return c
}

func collectForLevel(db *sql.DB, query string, perVarName *map[string]vars, level int) {
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		values := varObj{}
		var spare int
		rows.Scan(&values.id, &values.name, &values.peak, &spare, &values.value, &spare, &spare, &spare)
		//fmt.Printf("APP : %#v\n", values)
		v, ok := (*perVarName)[values.name]
		if !ok {
			v = vars{Levels: [4][]*varObj{[]*varObj{}, []*varObj{}, []*varObj{}, []*varObj{}}}
		}
		v.Levels[level] = append(v.Levels[level], &values)
		(*perVarName)[values.name] = v
	}

}

func collectAppVar(db *sql.DB, appName string) map[string]vars {

	perVarName := map[string]vars{}
	//AppOnly
	{
		query := "select vv.* from var as vv, (select distinct(v.id) from var as v ,application as a where v.appli=a.id and v.backendtype=0 and v.appsrv=0 and a.name='" + appName + "') as ee where ee.id=vv.id"
		collectForLevel(db, query, &perVarName, AppOnlyLevel)
	}

	//AppAs
	{
		query := "select vv.* from var as vv, (select distinct(v.id) from var as v ,application as a where v.backendtype=0 and v.appsrv>0 and a.id=v.appli and a.name='" + appName + "') as ee where ee.id=vv.id"
		collectForLevel(db, query, &perVarName, AppAsLevel)
	}

	//asBe
	{
		query := "select vv.* from var as vv, (select distinct(v.id) from var as v ,application as a, applicationserver as appsrv where v.appli=0 and v.backendtype>0 and v.appsrv=appsrv.id and a.id=appsrv.appli and a.name='" + appName + "') as ee where ee.id=vv.id"
		collectForLevel(db, query, &perVarName, AsBeLevel)
	}

	//BeOnly
	{
		query := "select vv.* from var as vv, (select distinct(v.id) from var as v ,application as a, applicationserver as appsrv , asbetype where v.appli=0 and v.appsrv=0 and v.backendtype=asbetype.betype and a.id=appsrv.appli and appsrv.id=asbetype.appsrv and a.name='" + appName + "') as ee where ee.id=vv.id"
		collectForLevel(db, query, &perVarName, BeOnlyLevel)
	}

	// ccc := 0
	// for _, v := range perVarName {
	// 	ccc += len(v.AppAsLevel)
	// 	ccc += len(v.AppOnlyLevel)
	// 	ccc += len(v.BeOnlyLevel)
	// 	ccc += len(v.AsBeLevel)
	// }
	//
	// fmt.Printf("Counts: %d", ccc)
	return perVarName
}

func printStatOnValuesForLevel(varname string, varsdata vars, level int, levelStr string, w *tabwriter.Writer) (int, int) {
	vals := map[string](struct{}){}
	savedCount := 0
	savedByte := 0

	for _, v := range varsdata.Levels[level] {
		vals[v.value] = struct{}{}
	}
	repeat := len(varsdata.Levels[level])
	if len(vals) <= 1 && repeat > 1 {
		varSize := len(varsdata.Levels[level][0].value)
		savedCount = savedCount + repeat - 1
		savedByte = savedByte + (repeat-1)*(varSize+len(varname))
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t\n", varname, levelStr, repeat, varSize)
	}

	return savedCount, savedByte
}

func printStatOnValues(data map[string]vars) (int, int, int) {

	savedCount := 0
	savedByte := 0

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 20, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintf(w, "Variable Name\tLevel\tRepetitions\tValue size\t\n")

	ptrPatternMatch := []string{}

	for varname, varsdata := range data {
		sc, sb := printStatOnValuesForLevel(varname, varsdata, AsBeLevel, "BE/AS", w)
		savedCount += sc
		savedByte += sb

		sc, sb = printStatOnValuesForLevel(varname, varsdata, BeOnlyLevel, "BE", w)
		savedCount += sc
		savedByte += sb

		sc, sb = printStatOnValuesForLevel(varname, varsdata, AppAsLevel, "AS", w)
		savedCount += sc
		savedByte += sb

		match, _ := regexp.Match("[0-9]{7}", []byte(varname))
		if match {
			ptrPatternMatch = append(ptrPatternMatch, varname)
		}
	}

	fmt.Fprintln(w)
	w.Flush()
	fmt.Printf("\nSaving %d definitions and %d bytes.\n", savedCount, savedByte)

	if len(ptrPatternMatch) > 0 {

		fmt.Printf("\nVariable to be checked (feature activation):\n")
		wexp := new(tabwriter.Writer)
		wexp.Init(os.Stdout, 20, 0, 2, ' ', tabwriter.AlignRight)
		fmt.Fprintf(wexp, "Variable Name\tRepetitions\t\n")
		for _, s := range ptrPatternMatch {
			v, _ := data[s]
			fmt.Fprintf(wexp, "%s\t%d\t\n", s, v.count())
		}
		fmt.Fprintln(wexp)
		wexp.Flush()
	}

	return savedCount, savedByte, len(ptrPatternMatch)
}
