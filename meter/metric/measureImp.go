package metric

import (
	"context"
	"github.com/cloudwego-contrib/cwgo-pkg/meter/label"
)

type MeasureImpl struct {
	counter  Counter
	recorder Recorder
}

func NewMeasure(counter Counter, recorder Recorder) Measure {
	return &MeasureImpl{
		counter:  counter,
		recorder: recorder,
	}
}

// Counter interface implementation
func (m *MeasureImpl) Inc(ctx context.Context, labels []label.CwLabel) error {
	return m.counter.Inc(ctx, labels)
}

func (m *MeasureImpl) Add(ctx context.Context, value int, labels []label.CwLabel) error {
	return m.counter.Add(ctx, value, labels)
}

// Recorder interface implementation
func (m *MeasureImpl) Record(ctx context.Context, value float64, labels []label.CwLabel) error {
	return m.recorder.Record(ctx, value, labels)
}
