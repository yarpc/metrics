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

// Package tallypush integrates go.uber.org/net/metrics with push-based StatsD
// and M3 systems.
package tallypush // import "go.uber.org/net/metrics/tallypush"

import (
	"math"
	"time"

	"github.com/uber-go/tally"
	"go.uber.org/net/metrics/push"
)

// New creates a push.Target that integrates with the Tally metrics package.
// Tally supports pushing to StatsD-based systems, M3, or both. See the Tally
// documentation for details: https://godoc.org/github.com/uber-go/tally.
func New(scope tally.Scope) push.Target {
	return &target{scope}
}

type target struct {
	tally.Scope
}

func (tp *target) NewCounter(spec push.Spec) push.Counter {
	return &counter{
		Counter: tp.Tagged(spec.Tags).Counter(spec.Name),
	}
}

func (tp *target) NewGauge(spec push.Spec) push.Gauge {
	return &gauge{tp.Tagged(spec.Tags).Gauge(spec.Name)}
}

func (tp *target) NewHistogram(spec push.HistogramSpec) push.Histogram {
	var buckets tally.Buckets
	if spec.Type == push.Duration {
		buckets = getDurationBuckets(spec.Buckets)
	} else {
		buckets = getValueBuckets(spec.Buckets) // This might normalize duration to "seconds" for all inputs, investigate.
	}
	return &histogram{
		Histogram: tp.Tagged(spec.Tags).Histogram(
			spec.Name,
			buckets,
		),
		lasts:         make(map[int64]int64, len(spec.Buckets)),
		histogramType: spec.Type,
	}
}

func getValueBuckets(buckets []int64) tally.Buckets {
	valueBuckets := make([]float64, len(buckets))
	for i := range buckets {
		if buckets[i] == math.MaxInt64 {
			valueBuckets[i] = math.MaxFloat64
		} else {
			valueBuckets[i] = float64(buckets[i])
		}
	}
	return tally.ValueBuckets(valueBuckets)
}

func getDurationBuckets(buckets []int64) tally.Buckets {
	durationBuckets := make([]time.Duration, len(buckets))
	for i := range buckets {
		if buckets[i] == math.MaxInt64 {
			durationBuckets[i] = time.Duration(math.MaxInt64)
		} else {
			durationBuckets[i] = time.Duration(buckets[i])
		}
	}
	return tally.DurationBuckets(durationBuckets)
}

type counter struct {
	tally.Counter

	last int64
}

func (c *counter) Set(total int64) {
	delta := total - c.last
	c.last = total
	c.Inc(delta)
}

type gauge struct {
	tally.Gauge
}

func (tg *gauge) Set(value int64) {
	tg.Update(float64(value))
}

type histogram struct {
	tally.Histogram

	// lasts keep the last value pushed to tally per histogram bucket.  This
	// defaults to zero.
	lasts map[int64]int64

	histogramType push.HistogramType
}

func (th *histogram) Set(bucket int64, total int64) {
	delta := total - th.lasts[bucket]
	th.lasts[bucket] = total

	for i := int64(0); i < delta; i++ {
		if th.histogramType == push.Duration {
			th.RecordDuration(time.Duration(bucket))
		} else {
			th.RecordValue(float64(bucket))
		}
	}
}
