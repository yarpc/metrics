// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package tallypush

import (
	"math"
	"testing"
	"time"

	"go.uber.org/net/metrics"
	"go.uber.org/net/metrics/push"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"
)

func newScope() tally.TestScope {
	return tally.NewTestScope("" /* prefix */, nil /* tags */)
}

func TestCounter(t *testing.T) {
	scope := newScope()
	target := New(scope)
	c := target.NewCounter(push.Spec{
		Name: "test_counter",
		Tags: metrics.Tags{"foo": "bar"},
	})
	c.Set(10)
	c.Set(20) // should overwrite previous value
	counters := scope.Snapshot().Counters()
	require.Equal(t, 1, len(counters), "Unexpected number of counters.")
	assert.Equal(t, int64(20), counters["test_counter+foo=bar"].Value(), "Unexpected exported value.")
}

func TestGauge(t *testing.T) {
	scope := newScope()
	target := New(scope)
	g := target.NewGauge(push.Spec{
		Name: "test_gauge",
		Tags: metrics.Tags{"foo": "bar"},
	})
	g.Set(10)
	g.Set(20) // should overwrite previous value
	gauges := scope.Snapshot().Gauges()
	require.Equal(t, 1, len(gauges), "Unexpected number of gauges.")
	assert.Equal(t, 20.0, gauges["test_gauge+foo=bar"].Value(), "Unexpected exported value.")
}

func TestValueHistogram(t *testing.T) {
	scope := newScope()
	target := New(scope)
	h := target.NewHistogram(push.HistogramSpec{
		Spec:    push.Spec{Name: "test_histogram", Tags: metrics.Tags{"foo": "bar"}},
		Buckets: []int64{5, 10, math.MaxInt64},
		Type:    push.Value,
	})
	h.Set(5, 1)
	h.Set(5, 2) // should overwrite previous value
	histograms := scope.Snapshot().Histograms()
	require.Equal(t, 1, len(histograms), "Unexpected number of histograms.")
	assert.Equal(
		t,
		map[float64]int64{5: 2, 10: 0, math.MaxFloat64: 0},
		histograms["test_histogram+foo=bar"].Values(),
	)
	assert.Nil(
		t,
		histograms["test_histogram+foo=bar"].Durations(),
	)
}

func TestDurationHistogram(t *testing.T) {
	scope := newScope()
	target := New(scope)
	h := target.NewHistogram(push.HistogramSpec{
		Spec:    push.Spec{Name: "test_histogram", Tags: metrics.Tags{"foo": "bar"}},
		Buckets: []int64{5, 10, math.MaxInt64},
		Type:    push.Duration,
	})
	h.Set(5, 1)
	h.Set(5, 2) // should overwrite previous value
	histograms := scope.Snapshot().Histograms()
	require.Equal(t, 1, len(histograms), "Unexpected number of histograms.")
	assert.Nil(
		t,
		histograms["test_histogram+foo=bar"].Values(),
	)
	assert.Equal(
		t,
		map[time.Duration]int64{time.Duration(5): 2, time.Duration(10): 0, time.Duration(math.MaxInt64): 0},
		histograms["test_histogram+foo=bar"].Durations(),
	)
}
