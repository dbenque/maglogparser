package log

import (
	"fmt"
	"maglogparser/locksearch/record"
	"maglogparser/locksearch/window"
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

	return
}
