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
	"sync"

	promproto "github.com/prometheus/client_model/go"
	"go.uber.org/atomic"
)

// Value is an atomic with some associated metadata. It's a building block
// for higher-level metric types.
type value struct {
	atomic.Int64

	meta       metadata
	labelPairs []*promproto.LabelPair
}

func newValue(m metadata) value {
	return value{
		meta:       m,
		labelPairs: m.MergeLabels(nil /* variable labels */),
	}
}

func newDynamicValue(m metadata, variableLabels []string) value {
	return value{
		meta:       m,
		labelPairs: m.MergeLabels(variableLabels),
	}
}

func (v value) snapshot() SimpleSnapshot {
	return SimpleSnapshot{
		Name:   *v.meta.Name,
		Labels: zip(v.labelPairs),
		Value:  v.Load(),
	}
}

// A vector is a collection of values that share the same metadata.
type vector struct {
	meta metadata

	// The factory function creates a new metric given the vector's metadata
	// and the variable label keys and values.
	factory func(metadata, []string) metric

	metricsMu sync.RWMutex
	metrics   map[string]metric // key is variable label vals
}

func (vec *vector) getOrCreate(labels ...string) (metric, error) {
	if err := vec.meta.ValidateVariableLabels(labels); err != nil {
		return nil, err
	}
	digester := newDigester()
	for i := 0; i < len(labels)/2; i++ {
		digester.add("", scrubLabelValue(labels[i*2+1]))
	}

	vec.metricsMu.RLock()
	m, ok := vec.metrics[string(digester.digest())]
	vec.metricsMu.RUnlock()
	if ok {
		digester.free()
		return m, nil
	}

	vec.metricsMu.Lock()
	m, err := vec.newValue(digester.digest(), labels)
	vec.metricsMu.Unlock()
	digester.free()

	return m, err
}

func (vec *vector) newValue(key []byte, variableLabels []string) (metric, error) {
	m, ok := vec.metrics[string(key)]
	if ok {
		return m, nil
	}
	m = vec.factory(vec.meta, variableLabels)
	vec.metrics[string(key)] = m
	return m, nil
}

func (vec *vector) snapshot() []SimpleSnapshot {
	vec.metricsMu.RLock()
	defer vec.metricsMu.RUnlock()
	snaps := make([]SimpleSnapshot, 0, len(vec.metrics))
	for _, m := range vec.metrics {
		switch v := m.(type) {
		case *Counter:
			snaps = append(snaps, v.snapshot())
		case *Gauge:
			snaps = append(snaps, v.snapshot())
		}
	}
	return snaps
}
