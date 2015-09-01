package record

import (
	"fmt"
	"log"
	"maglogparser/utils"
	"strings"
	"time"
)

type Record struct {
	Time                time.Time
	ThreadId            string
	Node                string
	Pid                 string
	Raw                 string
	Cmd                 string
	NextByThread        *Record
	PreviousByThread    *Record
	NextCmdRecord       *Record
	CmdRecord           *Record
	CmdCompletionRecord *Record
}

var RecordByThread map[string][]*Record

func init() {
	RecordByThread = make(map[string][]*Record)
}

func NewRecord(raw string) *Record {
	if len(raw) < 50 {
		return nil
	}
	if t, err := utils.ParseDate4(raw[0:len(utils.DateFormat)]); err == nil {
		r := Record{Time: t, Raw: raw}
		tokens := strings.Split(r.Raw, " ")
		r.Node = tokens[2]
		r.Pid = tokens[4]
		tokens = strings.Split(strings.Split(r.Raw, "<")[1], " ")
		r.ThreadId = tokens[1][4 : len(tokens[1])-1]

		if r.ThreadId[0] == 'f' {
			log.Fatalln(r.Raw)
		}

		return &r
	}
	return nil
}

// ScanThreadID place the received record in the good map per TID and chain the record previous/next
func ScanThreadID(c <-chan *Record) {

	for r := range c {
		m, ok := RecordByThread[r.ThreadId]
		if !ok {
			m = make([]*Record, 0, 0)
		}

		if len(m) > 0 {
			m[len(m)-1].NextByThread = r
			r.PreviousByThread = m[len(m)-1]
		}

		m = append(m, r)
		RecordByThread[r.ThreadId] = m
	}
}

func (r *Record) updateTextCommand() {

	if len(r.Cmd) > 0 {
		return
	}

	r.Cmd = strings.Split(strings.Split(r.Raw, "> ")[1], " ")[2]
	if strings.Contains(r.Cmd, "kSEIAdminCmdGeneric") {
		gen := strings.Split(strings.Split(r.Raw, "[")[1], "]")[0]
		tokens := strings.Split(gen, " ")
		if len(tokens) >= 5 {
			gen = tokens[3] + " " + tokens[4]
		}
		r.Cmd = r.Cmd + "[" + gen + "]"
	}
}

func (r *Record) IsCommand() bool {
	return strings.Contains(r.Raw, "> Processing Command#")
}

func (r *Record) IsEndCommand() bool {
	return strings.Contains(r.Raw, "> End of processing for Command#")
}

func (r *Record) GetCurrentCommand() *Record {

	wr := r

	for !wr.IsCommand() && wr.PreviousByThread != nil {
		wr = wr.PreviousByThread
	}

	if wr.IsCommand() {
		wr.updateTextCommand()
		r.CmdRecord = wr
		wr.CmdRecord = wr
		return wr
	}

	return nil
}

func (r *Record) GetNextCommand() *Record {

	if r.NextCmdRecord != nil {
		return r.NextCmdRecord
	}

	if r.NextByThread == nil {
		return nil
	}

	wr := r

	for wr.NextByThread != nil && !wr.NextByThread.IsCommand() {
		wr = wr.NextByThread
	}

	if wr.NextByThread == nil {
		return nil
	}

	r.NextCmdRecord = wr.NextByThread

	r.NextCmdRecord.updateTextCommand()

	return r.NextCmdRecord
}

func (r *Record) GetCommandCompletion() *Record {

	if r.CmdCompletionRecord != nil {
		return r.CmdCompletionRecord
	}

	wr := r

	for !wr.IsEndCommand() && wr.NextByThread != nil {
		wr = wr.NextByThread
	}

	if wr.NextByThread == nil {
		return nil
	}

	if wr.IsEndCommand() {
		r.CmdCompletionRecord = wr
		return wr
	}

	return nil
}

func (r *Record) GetCmdDuration() (time.Duration, error) {

	start := r.GetCurrentCommand()
	end := r.GetCommandCompletion()

	if start == nil || end == nil {
		return time.Duration(0), fmt.Errorf("Missing End or Start")
	}

	return end.Time.Sub(start.Time), nil

}
