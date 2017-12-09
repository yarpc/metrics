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

// A Scope is a collection of tagged metrics.
type Scope struct {
	core        *core
	constLabels Labels
}

func newScope(c *core, labels Labels) *Scope {
	return &Scope{
		core:        c,
		constLabels: labels,
	}
}

// Labeled creates a new Registry with new constant labels appended to the
// current Registry's existing labels. Label names and values are
// automatically scrubbed, with invalid characters replaced by underscores.
func (s *Scope) Labeled(ls Labels) *Scope {
	if s == nil {
		return nil
	}
	newLabels := make(Labels, len(s.constLabels)+len(ls))
	for k, v := range s.constLabels {
		newLabels[k] = v
	}
	for k, v := range ls {
		newLabels[scrubName(k)] = scrubLabelValue(v)
	}
	return newScope(s.core, newLabels)
}

// NewCounter constructs a new Counter.
func (s *Scope) NewCounter(spec Spec) (*Counter, error) {
	if s == nil {
		return nil, nil
	}
	spec = s.addConstLabels(spec)
	if err := spec.validateScalar(); err != nil {
		return nil, err
	}
	meta, err := newMetadata(spec)
	if err != nil {
		return nil, err
	}
	c := newCounter(meta)
	if err := s.core.register(c); err != nil {
		return nil, err
	}
	return c, nil
}

// NewGauge constructs a new Gauge.
func (s *Scope) NewGauge(spec Spec) (*Gauge, error) {
	if s == nil {
		return nil, nil
	}
	spec = s.addConstLabels(spec)
	if err := spec.validateScalar(); err != nil {
		return nil, err
	}
	meta, err := newMetadata(spec)
	if err != nil {
		return nil, err
	}
	g := newGauge(meta)
	if err := s.core.register(g); err != nil {
		return nil, err
	}
	return g, nil
}

// NewHistogram constructs a new Histogram.
func (s *Scope) NewHistogram(spec HistogramSpec) (*Histogram, error) {
	if s == nil {
		return nil, nil
	}
	spec.Spec = s.addConstLabels(spec.Spec)
	if err := spec.validateScalar(); err != nil {
		return nil, err
	}
	meta, err := newMetadata(spec.Spec)
	if err != nil {
		return nil, err
	}
	h := newHistogram(meta, spec.Unit, spec.Buckets)
	if err := s.core.register(h); err != nil {
		return nil, err
	}
	return h, nil
}

// NewCounterVector constructs a new CounterVector.
func (s *Scope) NewCounterVector(spec Spec) (*CounterVector, error) {
	if s == nil {
		return nil, nil
	}
	spec = s.addConstLabels(spec)
	if err := spec.validateVector(); err != nil {
		return nil, err
	}
	meta, err := newMetadata(spec)
	if err != nil {
		return nil, err
	}
	cv := newCounterVector(meta)
	if err := s.core.register(cv); err != nil {
		return nil, err
	}
	return cv, nil
}

// NewGaugeVector constructs a new GaugeVector.
func (s *Scope) NewGaugeVector(spec Spec) (*GaugeVector, error) {
	if s == nil {
		return nil, nil
	}
	spec = s.addConstLabels(spec)
	if err := spec.validateVector(); err != nil {
		return nil, err
	}
	meta, err := newMetadata(spec)
	if err != nil {
		return nil, err
	}
	gv := newGaugeVector(meta)
	if err := s.core.register(gv); err != nil {
		return nil, err
	}
	return gv, nil
}

// NewHistogramVector constructs a new HistogramVector.
func (s *Scope) NewHistogramVector(spec HistogramSpec) (*HistogramVector, error) {
	if s == nil {
		return nil, nil
	}
	spec.Spec = s.addConstLabels(spec.Spec)
	if err := spec.validateVector(); err != nil {
		return nil, err
	}
	meta, err := newMetadata(spec.Spec)
	if err != nil {
		return nil, err
	}
	hv := newHistogramVector(meta, spec.Unit, spec.Buckets)
	if err := s.core.register(hv); err != nil {
		return nil, err
	}
	return hv, nil
}

func (s *Scope) addConstLabels(spec Spec) Spec {
	if len(s.constLabels) == 0 {
		return spec
	}
	labels := make(Labels, len(s.constLabels)+len(spec.Labels))
	for k, v := range s.constLabels {
		labels[k] = v
	}
	for k, v := range spec.Labels {
		labels[k] = v
	}
	spec.Labels = labels
	return spec
}
