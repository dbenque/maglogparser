package record

import "strings"

type Records []*Record

func (a Records) Len() int      { return len(a) }
func (a Records) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type ByRecordTime struct{ Records }

func (a ByRecordTime) Less(i, j int) bool { return a.Records[i].Time.Before(a.Records[j].Time) }

type ByCmdName struct{ Records }

func (a ByCmdName) Less(i, j int) bool {
	return strings.Compare(a.Records[i].Cmd, a.Records[j].Cmd) < 0
	//return a[i].GetCmdName()[0] < a[j].GetCmdName()[0]
}

type ByDuration struct{ Records }

func (a ByDuration) Less(i, j int) bool {
	d1, e1 := a.Records[i].GetCmdDuration()
	d2, e2 := a.Records[j].GetCmdDuration()

	if e1 != nil && e2 != nil {
		return false
	}
	if e1 != nil {
		return false
	}

	return d1.Seconds() < d2.Seconds()
	//return a[i].GetCmdName()[0] < a[j].GetCmdName()[0]
}

type HasRecord interface {
	GetRecord() *Record
}
type HasRecords []HasRecord

func (a HasRecords) Len() int      { return len(a) }
func (a HasRecords) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type ByHasRecordTime struct{ HasRecords }

func (a ByHasRecordTime) Less(i, j int) bool {
	return a.HasRecords[i].GetRecord().Time.Before(a.HasRecords[j].GetRecord().Time)
}

// func strComp(a,b string) int {
//
// 	for i:=0; i<len(a) && i<len(b)
//
// }
