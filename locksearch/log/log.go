package log

import (
	"fmt"
	"maglogparser/locksearch/record"
	"maglogparser/locksearch/window"
	"maglogparser/utils"
)

func GetLog(tid string, count int) {

	records, ok := record.RecordByThread[tid]
	if !ok {
		fmt.Printf("Unknown TID: %s\n", tid)
		return
	}

	if len(records) == 0 {
		fmt.Printf("No records for the given thread\n", tid)
		return
	}

	t := window.GetWindow().Current

	index := 0
	if t.Before(records[0].Time) {
		index = 0
	} else if t.After(records[len(records)-1].Time) {
		index = len(records) - 1
	} else {

		for i, r := range records {
			if r.Time.After(t) || r.Time.Equal(t) {
				index = i
				break
			}
		}
	}

	if count == 0 {
		count = 5
	}

	if count > 0 {
		start := index - count
		end := index + count + 1
		if start < 0 {
			start = 0
		}
		if end > len(records) {
			end = len(records)
		}

		for i := start; i < end; i++ {
			prefix := "   "
			if t.Equal(records[i].Time) {
				prefix = " * "
			}
			fmt.Println(prefix + records[i].Raw)
		}
	} else {

		nCmd := records[index].GetCurrentCommand()

		if nCmd == nil {
			fmt.Printf("No command found for the given thread %s, at timestamp %s", tid, records[index].Time.Format(utils.DateFormat))
			return
		}

		for nCmd.NextByThread != nil && !nCmd.IsEndCommand() {
			fmt.Println(nCmd.Raw)
			nCmd = nCmd.NextByThread
		}

		if nCmd.IsEndCommand() {
			fmt.Println(nCmd.Raw)
		}
	}
	return
}
