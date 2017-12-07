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
	core        *coreRegistry
	constLabels Labels
}

func newRegistry(core *coreRegistry, labels Labels) *Registry {
	return &Registry{
		core:        core,
		constLabels: labels,
	}
}

// Labeled creates a new Registry with new constant labels appended to the
// current Registry's existing labels. Label names and values are
// automatically scrubbed, with invalid characters replaced by underscores.
func (r *Registry) Labeled(ls Labels) *Registry {
	if r == nil {
		return nil
	}
	newLabels := make(Labels, len(r.constLabels)+len(ls))
	for k, v := range r.constLabels {
		newLabels[k] = v
	}
	for k, v := range ls {
		newLabels[scrubName(k)] = scrubLabelValue(v)
	}
	return newRegistry(r.core, newLabels)
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
	meta, err := newMetadata(opts)
	if err != nil {
		return nil, err
	}
	c := newCounter(meta)
	if err := r.core.register(c); err != nil {
		return nil, err
	}
	return c, nil
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
	meta, err := newMetadata(opts)
	if err != nil {
		return nil, err
	}
	g := newGauge(meta)
	if err := r.core.register(g); err != nil {
		return nil, err
	}
	return g, nil
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
	meta, err := newMetadata(opts.Opts)
	if err != nil {
		return nil, err
	}
	h := newHistogram(meta)
	if err := r.core.register(h); err != nil {
		return nil, err
	}
	return h, nil
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
	meta, err := newMetadata(opts)
	if err != nil {
		return nil, err
	}
	cv := newCounterVector(meta)
	if err := r.core.register(cv); err != nil {
		return nil, err
	}
	return cv, nil
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
	meta, err := newMetadata(opts)
	if err != nil {
		return nil, err
	}
	gv := newGaugeVector(meta)
	if err := r.core.register(gv); err != nil {
		return nil, err
	}
	return gv, nil
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
	meta, err := newMetadata(opts.Opts)
	if err != nil {
		return nil, err
	}
	hv := newHistogramVector(meta)
	if err := r.core.register(hv); err != nil {
		return nil, err
	}
	return hv, nil
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
