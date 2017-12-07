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

package tallypush // import "go.uber.org/net/metrics/tallypush"

import (
	"math"

	"github.com/uber-go/tally"
	"go.uber.org/net/metrics/push"
)

// New creates a push.Target that integrates with the Tally metrics package.
// Tally supports pushing to StatsD-based systems, M3, or both - see the Tally
// documentation for details.
func New(scope tally.Scope) push.Target {
	return &target{scope}
}

type target struct {
	tally.Scope
}

func (tp *target) NewCounter(opts push.Opts) push.Counter {
	return &counter{
		Counter: tp.Tagged(opts.Labels).Counter(opts.Name),
	}
}

func (tp *target) NewGauge(opts push.Opts) push.Gauge {
	return &gauge{tp.Tagged(opts.Labels).Gauge(opts.Name)}
}

func (tp *target) NewHistogram(opts push.HistogramOpts) push.Histogram {
	buckets := make([]float64, len(opts.Buckets))
	for i := range opts.Buckets {
		if opts.Buckets[i] == math.MaxInt64 {
			buckets[i] = math.MaxFloat64
		} else {
			buckets[i] = float64(opts.Buckets[i])
		}
	}
	return &latency{
		Histogram: tp.Tagged(opts.Labels).Histogram(
			opts.Name,
			tally.ValueBuckets(buckets),
		),
		lasts: make(map[int64]int64, len(opts.Buckets)),
	}
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

type latency struct {
	tally.Histogram

	// lasts keep the last value pushed to tally per histogram bucket.  This
	// defaults to zero.
	lasts map[int64]int64
}

func (tg *latency) Set(bucket int64, total int64) {
	delta := total - tg.lasts[bucket]
	tg.lasts[bucket] = total

	for i := int64(0); i < delta; i++ {
		tg.RecordValue(float64(bucket))
	}
}
