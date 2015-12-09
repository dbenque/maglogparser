package main

import (
	"bufio"
	"fmt"
	"maglogparser/utils"
	"os"
	"time"
)

func main() {

	started := false
	t0 := time.Now()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < len(utils.DateFormat)+1 {
			continue
		}
		if t, err := utils.ParseDate5(line[0:len(utils.DateFormat)]); err == nil {

			if started {
				s := t.Sub(t0).Seconds()
				if s > 1.0 {
					fmt.Printf("Hole %fs between %s and %s\n", s, t0.Format(utils.DateFormat), t.Format(utils.DateFormat))
				}
			} else {
				started = true
			}

			t0 = t

		}

	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

}
