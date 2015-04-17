package magLogParserServer

import (
	"math/rand"
	"testing"
	"time"
)

type testTS struct {
	timestamp time.Time
}

func normalTPS(m metricContainer) {
	for i := 0; i < 60; i++ {
		for j := 0; j <= 100; j++ {
			m.AddMetric(&metric{time.Unix(int64(i+rand.Intn(3)), 0), 5})
		}
	}
}

func highTPS(m metricContainer) {
	for i := 0; i < 10; i++ {
		for j := 0; j <= 10000; j++ {
			m.AddMetric(&metric{time.Unix(int64(i+rand.Intn(3)), 0), 5})
		}
	}
}

func BenchmarkMetricArrayHighTPS(b *testing.B) {

	var metrics metricArray
	metrics.Init(60, 10)

	for i := 0; i < b.N; i++ {
		highTPS(&metrics)
	}

}

func BenchmarkMetricHeapHighTPS(b *testing.B) {

	var metrics metricHeap
	metrics.Init(60, 10)

	for i := 0; i < b.N; i++ {
		highTPS(&metrics)
	}

}

func BenchmarkMetricArrayNormal(b *testing.B) {

	var metrics metricArray
	metrics.Init(60, 10)

	for i := 0; i < b.N; i++ {
		normalTPS(&metrics)
	}

}

func BenchmarkMetricHeapNormalTPS(b *testing.B) {

	var metrics metricHeap
	metrics.Init(60, 10)

	for i := 0; i < b.N; i++ {
		normalTPS(&metrics)
	}

}

func (t *testTS) GetTimestamp() time.Time {
	return t.timestamp
}

func (t *testTS) String() string { return t.timestamp.String() }

// func TestUpdateListUpdate(T *testing.T) {
//
// 	L := list.New()
// 	t := time.Now()
// 	t0 := testTS{t}
// 	t1 := testTS{t.Add(time.Second)}
//
// 	updateListUpdate(L, 4, &t0)
// 	updateListUpdate(L, 4, &t1)
//
// 	if L.Len() != 2 {
// 		T.Fatal("Check Failed")
// 	}
//
// 	t2 := testTS{t.Add(time.Second * 4)}
// 	updateListUpdate(L, 4, &t2)
//
// 	if L.Len() != 3 {
// 		T.Fatal("Check Failed")
// 	}
//
// 	t3 := testTS{t.Add(time.Second)}
// 	updateListUpdate(L, 4, &t3)
//
// 	if L.Len() != 3 {
// 		T.Fatal("Check Failed")
// 	}
//
// 	t4 := testTS{t.Add(time.Second * 2)}
// 	updateListUpdate(L, 4, &t4)
//
// 	if L.Len() != 4 {
// 		T.Fatal("Check Failed")
// 	}
//
// 	t5 := testTS{t}
// 	updateListUpdate(L, 4, &t5)
//
// 	if L.Len() != 4 {
// 		T.Fatal("Check Failed")
// 	}
//
// 	t6 := testTS{t.Add(time.Second * 4)}
// 	updateListUpdate(L, 4, &t6)
//
// 	if L.Len() != 4 {
// 		T.Fatal("Check Failed")
// 	}
//
// 	e := L.Back()
// 	ts, _ := e.Value.(timestamped)
// 	if !ts.GetTimestamp().Equal(t.Add(time.Second * 4)) {
// 		T.Fatal("Check Failed")
// 	}
//
// 	e = e.Prev()
// 	ts, _ = e.Value.(timestamped)
// 	if !ts.GetTimestamp().Equal(t.Add(time.Second * 2)) {
// 		T.Fatal("Check Failed")
// 	}
//
// 	e = e.Prev()
// 	ts, _ = e.Value.(timestamped)
// 	if !ts.GetTimestamp().Equal(t.Add(time.Second)) {
// 		T.Fatal("Check Failed")
// 	}
//
// 	e = e.Prev()
// 	ts, _ = e.Value.(timestamped)
// 	if !ts.GetTimestamp().Equal(t) {
// 		T.Fatal("Check Failed")
// 	}
//
// 	t7 := testTS{t.Add(time.Second * 3)}
// 	updateListUpdate(L, 4, &t7)
// 	if L.Len() != 4 {
// 		T.Fatal("Check Failed")
// 	}
//
// 	e = L.Front()
// 	ts, _ = e.Value.(timestamped)
// 	if !ts.GetTimestamp().Equal(t.Add(time.Second)) {
// 		T.Fatal("Check Failed")
// 	}
//
// }
