package main

import "maglogparser/locksearch/record"

type HasRecord interface {
	GetRecord() *record.Record
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
	//return strings.Compare(a[i].GetCmdName(), a[j].GetCmdName()) < 0
	return a[i].GetCmdName()[0] < a[j].GetCmdName()[0]
}

// func strComp(a,b string) int {
//
// 	for i:=0; i<len(a) && i<len(b)
//
// }
