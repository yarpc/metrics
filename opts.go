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
	"errors"
	"fmt"
	"math"
	"time"
)

// Opts configure Counters, Gauges, CounterVectors, and GaugeVectors.
type Opts struct {
	Name           string
	Help           string
	Labels         Labels
	VariableLabels []string // only meaningful for vectors
	DisablePush    bool
}

func (o Opts) validate() error {
	if len(o.Name) == 0 {
		return errors.New("all metrics must have a name")
	}
	if o.Help == "" {
		return errors.New("metric help must not be empty")
	}
	return nil
}

func (o Opts) validateScalar() error {
	if err := o.validate(); err != nil {
		return err
	}
	if len(o.VariableLabels) > 0 {
		return errors.New("only vectors may have variable labels")
	}
	return nil
}

func (o Opts) validateVector() error {
	if err := o.validate(); err != nil {
		return err
	}
	if len(o.VariableLabels) == 0 {
		return errors.New("vectors must have variable labels")
	}
	return nil
}

// HistogramOpts configure Histograms and HistogramVectors.
type HistogramOpts struct {
	Opts

	// Durations are exported to Prometheus as simple numbers, not strings or
	// rich objects. Unit specifies the desired granularity for latency
	// observations. For example, an observation of time.Second with a unit of
	// time.Millisecond is exported to Prometheus as 1000. Typically, the unit
	// should also be part of the metric name; in this example, latency_ms is a
	// good name.
	Unit time.Duration
	// Upper bounds (inclusive) for the histogram buckets. A catch-all bucket
	// for large observations is automatically created, if necessary.
	Buckets []int64
}

func (ho HistogramOpts) validateScalar() error {
	if err := ho.validateLatencies(); err != nil {
		return err
	}
	return ho.Opts.validateScalar()
}

func (ho HistogramOpts) validateVector() error {
	if err := ho.validateLatencies(); err != nil {
		return err
	}
	return ho.Opts.validateVector()
}

func (ho HistogramOpts) validateLatencies() error {
	if ho.Unit < 1 {
		return fmt.Errorf("duration unit must be positive, got %v", ho.Unit)
	}
	if len(ho.Buckets) == 0 {
		return fmt.Errorf("must specify some buckets")
	}
	prev := int64(math.MinInt64)
	for _, upper := range ho.Buckets {
		if upper <= prev {
			return fmt.Errorf("bucket upper bounds must be sorted in increasing order")
		}
		prev = upper
	}
	return nil
}
