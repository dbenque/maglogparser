package magLogParserServer

import (
	"container/heap"
)

// An IntHeap is a min-heap of ints.
type metricHeap struct {
	values      []metricSummary
	updateCount uint
	size        uint
}

// ----------------------- Implement Metric Container interface ---------

func (h *metricHeap) AddMetric(m *metric) {

	// As the latest input are probably linked to more recent item, search from back in the heap
	for i := len(h.values) - 1; i >= 0; i-- {

		// Item found, update summary
		if h.values[i].Timestamp.Equal(m.timestamp) {
			h.values[i].Update(m)
			return
		}

		if m.timestamp.Before(h.values[i].Timestamp) {
			continue
		} else {
			// there are some older insert in the list: need to update
			ms := new(metricSummary)
			ms.Update(m)
			heap.Push(h, *ms)

			// Check if we have not exceeded the size
			if uint(len(h.values)) > h.size {
				heap.Pop(h)
			}
			return
		}
	}

	// is the item to old to enter the heap, at this stage it means that we have not free slot up into the heap to insert it
	if uint(len(h.values)) < h.size {
		ms := new(metricSummary)
		ms.Update(m)
		heap.Push(h, *ms)
	}
}
func (h *metricHeap) Init(size uint, updates uint) {
	h.values = make([]metricSummary, 0, 0)
	h.updateCount = updates
	h.size = size
}

func (h *metricHeap) GetAllMetrics() []metricSummary {
	return h.values
}

// -------------------- Implement Sort Interface ------------------------

func (h metricHeap) Len() int { return len(h.values) }
func (h metricHeap) Less(i, j int) bool {
	return h.values[i].GetTimestamp().Before(h.values[j].GetTimestamp())
}
func (h metricHeap) Swap(i, j int) { h.values[i], h.values[j] = h.values[j], h.values[i] }

// -------------------- Implement Heap Interface ------------------------

func (h *metricHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	h.values = append(h.values, x.(metricSummary))
}

func (h *metricHeap) Pop() interface{} {
	old := h.values
	n := len(old)
	x := old[n-1]
	h.values = old[0 : n-1]
	return x
}
