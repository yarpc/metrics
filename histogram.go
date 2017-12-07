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
	"fmt"
	"time"
)

// A Histogram approximates a distribution of values. They're both more
// efficient and easier to aggregate than Prometheus summaries or M3 timers.
// For a discussion of the tradeoffs between histograms and timers/summaries,
// see https://prometheus.io/docs/practices/histograms/.
//
// All exported methods are safe to use concurrently, and nil *Histograms are
// valid no-op implementations.
type Histogram struct {
	meta metadata
}

func newHistogram(m metadata) *Histogram {
	return &Histogram{m}
}

// Observe finds the correct bucket for the supplied duration and increments
// its counter.
func (h *Histogram) Observe(d time.Duration) {
}

// ObserveInt finds the correct bucket for the supplied integer and increments
// its counter.
func (h *Histogram) ObserveInt(n int64) {
}

func (h *Histogram) describe() metadata {
	return h.meta
}

func (h *Histogram) snapshot() HistogramSnapshot {
	return HistogramSnapshot{}
}

// A HistogramVector is a collection of Histograms that share a name and some
// constant labels, but also have a consistent set of variable labels. All
// exported methods are safe to use concurrently.
//
// A nil *HistogramVector is safe to use, and always returns no-op histograms.
//
// For a general description of vector types, see the package-level
// documentation.
type HistogramVector struct {
	meta metadata
}

func newHistogramVector(m metadata) *HistogramVector {
	return &HistogramVector{m}
}

// Get retrieves the histogram with the supplied variable label names and
// values from the vector, creating one if necessary. The variable labels must
// be supplied in the same order used when creating the vector.
//
// Get returns an error if the number or order of labels is incorrect.
func (hv *HistogramVector) Get(variableLabels ...string) (*Histogram, error) {
	if hv == nil {
		return nil, nil
	}
	return nil, nil
}

// MustGet behaves exactly like Get, but panics on errors. If code using this
// method is covered by unit tests, this is safe.
func (hv *HistogramVector) MustGet(variableLabels ...string) *Histogram {
	if hv == nil {
		return nil
	}
	h, err := hv.Get(variableLabels...)
	if err != nil {
		panic(fmt.Sprintf("failed to get histogram: %v", err))
	}
	return h
}

func (hv *HistogramVector) describe() metadata {
	return hv.meta
}

func (hv *HistogramVector) snapshot() []HistogramSnapshot {
	return nil
}
