package magLogParserServer

import "time"

type timestamped interface {
	GetTimestamp() time.Time
}

type metric struct {
	timestamp time.Time
	value     int
}

type metricSummary struct {
	Timestamp time.Time
	Max       int
	Min       int
	Count     int
	Sum       int64
}

func (mSummary *metricSummary) GetTimestamp() time.Time {
	return mSummary.Timestamp
}

func (mSummary *metricSummary) Update(m *metric) {
	// update the summary
	if mSummary.Timestamp.IsZero() {
		mSummary.Max = m.value
		mSummary.Min = m.value
	}

	mSummary.Timestamp = m.timestamp

	if mSummary.Max < m.value {
		mSummary.Max = m.value
	}

	if mSummary.Min > m.value {
		mSummary.Min = m.value
	}

	mSummary.Sum += int64(m.value)
	mSummary.Count++
}

type metricContainer interface {
	// Add a Metric in the container
	AddMetric(m *metric)
	// Initialize the Container with Max collection size (returned by GetAllMetrics) and the latest updated records(returned by GetLatestMetrics)
	Init(size uint, updates uint)
	// Return all the records
	GetAllMetrics() []metricSummary
}
