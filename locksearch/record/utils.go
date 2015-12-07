package record

import (
	"strings"
	"time"
)

type HasRecord interface {
	GetRecord() *Record
}

type ByRecordTime []HasRecord

func (a ByRecordTime) Len() int           { return len(a) }
func (a ByRecordTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByRecordTime) Less(i, j int) bool { return a[i].GetRecord().Time.Before(a[j].GetRecord().Time) }

type HasCmdName interface {
	GetCmdName() string
}

type ByCmdName []HasCmdName

func (a ByCmdName) Len() int      { return len(a) }
func (a ByCmdName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByCmdName) Less(i, j int) bool {
	return strings.Compare(a[i].GetCmdName(), a[j].GetCmdName()) < 0
	//return a[i].GetCmdName()[0] < a[j].GetCmdName()[0]
}

type HasDuration interface {
	GetCmdDuration() (time.Duration, error)
}

type ByDuration []HasDuration

func (a ByDuration) Len() int      { return len(a) }
func (a ByDuration) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDuration) Less(i, j int) bool {
	d1, e1 := a[i].GetCmdDuration()
	d2, e2 := a[i].GetCmdDuration()

	if e1 != nil && e2 != nil {
		return false
	}
	if e1 != nil {
		return false
	}

	return d1.Seconds() < d2.Seconds()
	//return a[i].GetCmdName()[0] < a[j].GetCmdName()[0]
}

// func strComp(a,b string) int {
//
// 	for i:=0; i<len(a) && i<len(b)
//
// }
