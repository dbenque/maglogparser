package magLogParserServer

import "sort"

type metricSummaries []metricSummary

type metricArray struct {
	size        uint64
	values      metricSummaries
	updateCount uint
	shift       uint // Used to start the slice at index 0
	maxIndex    uint // What will be the end of the slice
}

func (h metricSummaries) Len() int { return len(h) }
func (h metricSummaries) Less(i, j int) bool {
	return h[i].GetTimestamp().Before(h[j].GetTimestamp())
}
func (h metricSummaries) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (mArray *metricArray) GetAllMetrics() []metricSummary {
	local := make(metricSummaries, len(mArray.values), len(mArray.values))
	copy(local, mArray.values)
	sort.Sort(local)
	return local
}

func (mArray *metricArray) Init(size uint, updateCount uint) {
	mArray.size = uint64(size)
	mArray.values = make([]metricSummary, size+1, size+1)
	mArray.updateCount = updateCount
	mArray.maxIndex = 0
}

func (mArray *metricArray) AddMetric(m *metric) {

	t := uint64(m.timestamp.Unix())

	index := t % mArray.size

	summary := &(mArray.values[index])
	if summary.Timestamp.Equal(m.timestamp) {
		summary.Update(m)
	} else {
		ms := new(metricSummary)
		ms.Update(m)
		mArray.values[index] = *ms
	}

}
