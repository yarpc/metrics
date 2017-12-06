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

// A Registry is a scoped metric registry. All metrics created with a Registry
// will have the Registry's labels appended to them.
type Registry struct {
	constLabels Labels
}

func newRegistry(labels Labels) *Registry {
	return &Registry{
		constLabels: labels,
	}
}

// NewCounter constructs a new Counter.
func (r *Registry) NewCounter(opts Opts) (*Counter, error) {
	if r == nil {
		return nil, nil
	}
	opts = r.addConstLabels(opts)
	if err := opts.validateScalar(); err != nil {
		return nil, err
	}
	return nil, nil
}

// NewGauge constructs a new Gauge.
func (r *Registry) NewGauge(opts Opts) (*Gauge, error) {
	if r == nil {
		return nil, nil
	}
	opts = r.addConstLabels(opts)
	if err := opts.validateScalar(); err != nil {
		return nil, err
	}
	return nil, nil
}

// NewHistogram constructs a new Histogram.
func (r *Registry) NewHistogram(opts HistogramOpts) (*Histogram, error) {
	if r == nil {
		return nil, nil
	}
	opts.Opts = r.addConstLabels(opts.Opts)
	if err := opts.validateScalar(); err != nil {
		return nil, err
	}
	return nil, nil
}

// NewCounterVector constructs a new CounterVector.
func (r *Registry) NewCounterVector(opts Opts) (*CounterVector, error) {
	if r == nil {
		return nil, nil
	}
	opts = r.addConstLabels(opts)
	if err := opts.validateVector(); err != nil {
		return nil, err
	}
	return nil, nil
}

// NewGaugeVector constructs a new GaugeVector.
func (r *Registry) NewGaugeVector(opts Opts) (*GaugeVector, error) {
	if r == nil {
		return nil, nil
	}
	opts = r.addConstLabels(opts)
	if err := opts.validateVector(); err != nil {
		return nil, err
	}
	return nil, nil
}

// NewHistogramVector constructs a new HistogramVector.
func (r *Registry) NewHistogramVector(opts HistogramOpts) (*HistogramVector, error) {
	if r == nil {
		return nil, nil
	}
	opts.Opts = r.addConstLabels(opts.Opts)
	if err := opts.validateVector(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *Registry) addConstLabels(opts Opts) Opts {
	if len(r.constLabels) == 0 {
		return opts
	}
	labels := make(Labels, len(r.constLabels)+len(opts.Labels))
	for k, v := range r.constLabels {
		labels[k] = v
	}
	for k, v := range opts.Labels {
		labels[k] = v
	}
	opts.Labels = labels
	return opts
}
