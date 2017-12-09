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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGauge(t *testing.T) {
	root := New()
	s := root.Scope().Labeled(Labels{"service": "users"})

	t.Run("duplicate constant labels", func(t *testing.T) {
		_, err := s.NewGauge(Spec{
			Name:   "test_gauge",
			Help:   "help",
			Labels: Labels{"f_": "ok", "f&": "ok"}, // scrubbing introduces duplicate label names
		})
		assert.Error(t, err, "Expected an error constructing a gauge with invalid spec.")
	})

	t.Run("valid spec", func(t *testing.T) {
		gauge, err := s.NewGauge(Spec{
			Name:   "test_gauge",
			Help:   "Some help.",
			Labels: Labels{"foo": "bar"},
		})
		require.NoError(t, err, "Unexpected error constructing gauge.")

		assert.Equal(t, int64(1), gauge.Inc(), "Unexpected return value from increment.")
		assert.Equal(t, int64(0), gauge.Dec(), "Unexpected return value from decrement.")
		assert.Equal(t, int64(0), gauge.Swap(1), "Unexpected return value from swap.")

		gauge.Store(42)
		assert.Equal(t, int64(42), gauge.Load(), "Unexpected in-memory gauge value.")

		assert.True(t, gauge.CAS(42, 43), "Unexpected return value from CAS.")
		snap := root.Snapshot()
		require.Equal(t, 1, len(snap.Gauges), "Unexpected number of gauges.")
		assert.Equal(t, Snapshot{
			Name:   "test_gauge",
			Labels: Labels{"foo": "bar", "service": "users"},
			Value:  43,
		}, snap.Gauges[0], "Unexpected gauge snapshot.")
	})
}

func TestGaugeVector(t *testing.T) {
	newVector := func() (*GaugeVector, *Root) {
		root := New()
		spec := Spec{
			Name:           "test_gauge",
			Help:           "Some help.",
			VariableLabels: []string{"var"},
		}
		vec, err := root.Scope().NewGaugeVector(spec)
		require.NoError(t, err, "Unexpected error constructing vector.")
		return vec, root
	}

	assertGauge := func(root *Root, expectedLabel string, expectedReading int64) {
		snap := root.Snapshot()
		require.Equal(t, 1, len(snap.Gauges), "Unexpected number of gauges.")
		got := snap.Gauges[0]
		assert.Equal(t, "test_gauge", got.Name, "Unexpected name.")
		assert.Equal(t, Labels{"var": expectedLabel}, got.Labels, "Unexpected labels.")
		assert.Equal(t, expectedReading, got.Value, "Unexpected gauge value.")
	}

	t.Run("valid labels", func(t *testing.T) {
		vec, root := newVector()
		g, err := vec.Get("var", "x")
		require.NoError(t, err, "Unexpected error getting gauge.")

		g.Store(1)
		vec.MustGet("var", "x").Add(2)

		assertGauge(root, "x", 3)
	})

	t.Run("invalid labels", func(t *testing.T) {
		vec, root := newVector()
		g, err := vec.Get("var", "x!")
		require.NoError(t, err, "Unexpected error getting gauge.")

		g.Store(1)
		vec.MustGet("var", "x!").Inc()
		vec.MustGet("var", "x&").Inc()

		assertGauge(root, "x_", 3)
	})

	t.Run("cardinality mismatch", func(t *testing.T) {
		vec, _ := newVector()
		_, err := vec.Get("var", "x", "var2", "y")
		assert.Error(t, err, "Expected an error getting a gauge with too many labels.")
		assert.Panics(t, func() {
			vec.MustGet("var", "x", "var2", "y")
		}, "Expected a panic using MustGet with the wrong number of labels.")
	})
}

func TestGaugeVectorConstructionErrors(t *testing.T) {
	s := New().Scope()

	t.Run("duplicate constant label names", func(t *testing.T) {
		_, err := s.NewGaugeVector(Spec{
			Name:           "test_gauge",
			Help:           "help",
			Labels:         Labels{"f_": "ok", "f&": "ok"}, // scrubbing introduces duplicate label names
			VariableLabels: []string{"var"},
		})
		assert.Error(t, err, "Expected an error constructing a gauge vector with invalid spec.")
	})

	t.Run("duplicate variable label names", func(t *testing.T) {
		_, err := s.NewGaugeVector(Spec{
			Name:           "test_gauge",
			Help:           "help",
			VariableLabels: []string{"var", "var"},
		})
		assert.Error(t, err, "Expected an error constructing a gauge vector with invalid spec.")
	})
}
