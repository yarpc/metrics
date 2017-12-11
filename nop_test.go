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

package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNopCounter(t *testing.T) {
	assertNopCounter(t, nil)
}

func TestNopCounterVector(t *testing.T) {
	assertNopCounterVector(t, nil)
}

func TestNopGauge(t *testing.T) {
	assertNopGauge(t, nil)
}

func TestNopGaugeVector(t *testing.T) {
	assertNopGaugeVector(t, nil)
}

func TestNopHistogram(t *testing.T) {
	assertNopHistogram(t, nil)
}

func TestNopHistogramVector(t *testing.T) {
	assertNopHistogramVector(t, nil)
}

func TestNopScope(t *testing.T) {
	var s *Scope
	s = s.Tagged(Tags{"foo": "bar"})
	c, err := s.Counter(Spec{})
	assert.NoError(t, err, "Error calling Counter on nil scope.")
	assertNopCounter(t, c)

	cv, err := s.CounterVector(Spec{})
	assert.NoError(t, err, "Error calling CounterVector on nil scope.")
	assertNopCounterVector(t, cv)

	g, err := s.Gauge(Spec{})
	assert.NoError(t, err, "Error calling Gauge on nil scope.")
	assertNopGauge(t, g)

	gv, err := s.GaugeVector(Spec{})
	assert.NoError(t, err, "Error calling GaugeVector on nil scope.")
	assertNopGaugeVector(t, gv)

	h, err := s.Histogram(HistogramSpec{})
	assert.NoError(t, err, "Error calling Histogram on nil scope.")
	assertNopHistogram(t, h)

	hv, err := s.HistogramVector(HistogramSpec{})
	assertNopHistogramVector(t, hv)
}

func assertNopCounter(t testing.TB, c *Counter) {
	assert.Equal(t, int64(0), c.Add(42), "Unexpected result from no-op Add.")
	assert.Equal(t, int64(0), c.Inc(), "Unexpected result from no-op Inc.")
	assert.Equal(t, int64(0), c.Load(), "Unexpected result from no-op Load.")
}

func assertNopCounterVector(t *testing.T, vec *CounterVector) {
	c, err := vec.Get("foo", "bar")
	require.NoError(t, err, "Failed Get from no-op CounterVector.")
	assert.NotPanics(t, func() {
		vec.MustGet("foo", "bar")
	}, "Failed MustGet from no-op CounterVector.")
	assertNopCounter(t, c)
}

func assertNopGauge(t testing.TB, g *Gauge) {
	g.Store(42)
	assert.Equal(t, int64(0), g.Add(42), "Unexpected result from no-op Add.")
	assert.Equal(t, int64(0), g.Sub(1), "Unexpected result from no-op Sub.")
	assert.Equal(t, int64(0), g.Inc(), "Unexpected result from no-op Inc.")
	assert.Equal(t, int64(0), g.Dec(), "Unexpected result from no-op Dec.")
	assert.Equal(t, int64(0), g.Load(), "Unexpected result from no-op Load.")
	assert.Equal(t, int64(0), g.Swap(42), "Unexpected result from no-op Swap.")
	assert.True(t, g.CAS(42, 10), "Unexpected result from no-op CAS.")
}

func assertNopGaugeVector(t testing.TB, vec *GaugeVector) {
	g, err := vec.Get("foo", "bar")
	require.NoError(t, err, "Failed Get from no-op GaugeVector.")
	assert.NotPanics(t, func() {
		vec.MustGet("foo", "bar")
	}, "Failed MustGet from no-op GaugeVector.")
	assertNopGauge(t, g)
}

func assertNopHistogram(t testing.TB, h *Histogram) {
	assert.NotPanics(t, func() {
		h.Observe(time.Second)
		h.IncBucket(42)
	}, "Unexpected panic using no-op histgram.")
}

func assertNopHistogramVector(t testing.TB, vec *HistogramVector) {
	h, err := vec.Get("foo", "bar")
	require.NoError(t, err, "Failed Get from no-op HistogramVector.")
	assert.NotPanics(t, func() {
		vec.MustGet("foo", "bar")
	}, "Failed MustGet from no-op HistogramVector.")
	assertNopHistogram(t, h)
}
