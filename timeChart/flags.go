package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/dbenque/maglogparser/utils"
)

var timeFormatFlag string
var periodUnitFlag string
var periodValueFlag uint
var serieCountFlag uint
var fileFlag string
var noMaxFlag bool
var noMinFlag bool
var noAvgFlag bool

var mandatoryFlags, optionalFlags *flag.FlagSet

func InitFlags() {
	mandatoryFlags = flag.NewFlagSet("Manadatory Flags", flag.ContinueOnError)
	mandatoryFlags.StringVar(&timeFormatFlag, "t", utils.DateFormat, "Time Format")
	mandatoryFlags.StringVar(&fileFlag, "f", "", "file to read")

	optionalFlags = flag.NewFlagSet("Optional Flags", flag.ContinueOnError)
	optionalFlags.StringVar(&periodUnitFlag, "u", "s", "Unit of the period")
	optionalFlags.UintVar(&periodValueFlag, "p", 1, "Value of the period")
	optionalFlags.UintVar(&serieCountFlag, "c", 1, "Number of series")
	optionalFlags.BoolVar(&noMaxFlag, "noMax", false, "Don't display Max")
	optionalFlags.BoolVar(&noMaxFlag, "noMin", false, "Don't display Min")
	optionalFlags.BoolVar(&noAvgFlag, "noAvg", false, "Don't display Avg")

	flag.Usage = FlagUsage
}

func FlagUsage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	mandatoryFlags.SetOutput(os.Stderr)
	optionalFlags.SetOutput(os.Stderr)
	fmt.Fprintf(os.Stderr, "\nMandatory parameters:\n")
	mandatoryFlags.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nOptional parameters:\n")
	optionalFlags.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExample:\n")
	fmt.Fprintf(os.Stderr, "\n%s -f=kSEIAgentConfigRequested.mavg -t=20060102 -p=1 -u=day\n\n", os.Args[0])
}

func ParseFlags() error {

	// Parsing all FlagSet
	for i := range os.Args[1:] {
		buf := new(bytes.Buffer)
		mandatoryFlags.SetOutput(buf)
		optionalFlags.SetOutput(buf)
		if err := mandatoryFlags.Parse(os.Args[i+1 : i+2]); err == nil {
			continue
		}
		if err := optionalFlags.Parse(os.Args[i+1 : i+2]); err == nil {
			continue
		}
		eflag := os.Args[i+1]
		switch eflag {
		case "--help", "-help", "-h", "--h", "-usage", "--usage":
			flag.Usage()
			return fmt.Errorf("Help requested")
		default:
			fmt.Fprintf(os.Stderr, "Error with parameter: %s\n", eflag)
			flag.Usage()
			return fmt.Errorf("Unknown flag or missing value")

		}
	}

	// Checking Mandatory flags
	mandatoryFlagsParsing := map[string]*flag.Flag{}
	mandatoryFlags.VisitAll(func(f *flag.Flag) {
		mandatoryFlagsParsing[f.Name] = f
	})
	mandatoryFlags.Visit(func(f *flag.Flag) {
		delete(mandatoryFlagsParsing, f.Name)
	})

	if len(mandatoryFlagsParsing) != 0 {
		for _, f := range mandatoryFlagsParsing {
			fmt.Fprintf(os.Stderr, "Missing mandatory flag: %s\n", f.Name)
		}
		FlagUsage()
		return fmt.Errorf("Missing mandatory flag")
	}
	return nil
}
