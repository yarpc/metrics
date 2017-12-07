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

func TestCounter(t *testing.T) {
	r, c := New()

	t.Run("duplicate constant label names", func(t *testing.T) {
		_, err := r.NewCounter(Opts{
			Name:   "test_counter",
			Help:   "help",
			Labels: Labels{"f_": "ok", "f&": "ok"}, // scrubbing introduces duplicate label names
		})
		assert.Error(t, err, "Expected an error constructing a counter with invalid options.")
	})

	t.Run("valid opts", func(t *testing.T) {
		counter, err := r.NewCounter(Opts{
			Name:   "test_counter",
			Help:   "Some help.",
			Labels: Labels{"foo": "bar"},
		})
		require.NoError(t, err, "Unexpected error constructing counter.")

		assert.Equal(t, int64(1), counter.Inc(), "Unexpected return value from increment.")
		assert.Equal(t, int64(3), counter.Add(2), "Unexpected return value from add.")
		assert.Equal(t, int64(3), counter.Add(-1), "Should forbid decrementing counters.")
		assert.Equal(t, int64(3), counter.Load(), "Unexpected in-memory counter value.")

		snap := c.Snapshot()
		require.Equal(t, 1, len(snap.Counters), "Unexpected number of counters.")
		assert.Equal(t, SimpleSnapshot{
			Name:   "test_counter",
			Labels: Labels{"foo": "bar"},
			Value:  3,
		}, snap.Counters[0], "Unexpected counter snapshot.")
	})
}

func TestCounterVector(t *testing.T) {
	newVector := func() (*CounterVector, *Controller) {
		r, c := New()
		opts := Opts{
			Name:           "test_counter",
			Help:           "Some help.",
			VariableLabels: []string{"var"},
		}
		vec, err := r.NewCounterVector(opts)
		require.NoError(t, err, "Unexpected error constructing vector.")
		return vec, c
	}

	assertCounter := func(c *Controller, expectedLabel string, expectedCount int64) {
		snap := c.Snapshot()
		require.Equal(t, 1, len(snap.Counters), "Unexpected number of counters.")
		got := snap.Counters[0]
		assert.Equal(t, SimpleSnapshot{
			Name:   "test_counter",
			Labels: Labels{"var": expectedLabel},
			Value:  expectedCount,
		}, got, "Unexpected counter snapshot.")
	}

	t.Run("valid labels", func(t *testing.T) {
		vec, c := newVector()
		counter, err := vec.Get("var", "x")
		require.NoError(t, err, "Unexpected error getting counter.")

		counter.Inc()
		vec.MustGet("var", "x").Add(2)

		assertCounter(c, "x", 3)
	})

	t.Run("invalid labels", func(t *testing.T) {
		vec, c := newVector()
		counter, err := vec.Get("var", "x!")
		require.NoError(t, err, "Unexpected error getting counter.")

		counter.Inc()
		vec.MustGet("var", "x!").Inc()
		vec.MustGet("var", "x&").Inc()

		assertCounter(c, "x_", 3)
	})

	t.Run("cardinality mismatch", func(t *testing.T) {
		vec, _ := newVector()
		_, err := vec.Get("var", "x", "var2", "y")
		assert.Error(t, err, "Expected an error getting a counter with too many labels.")
		assert.Panics(t, func() {
			vec.MustGet("var", "x", "var2", "y")
		}, "Expected a panic using MustGet with the wrong number of labels.")
	})
}

func TestCounterVectorConstructionErrors(t *testing.T) {
	r, _ := New()

	t.Run("duplicate constant label names", func(t *testing.T) {
		_, err := r.NewCounterVector(Opts{
			Name:           "test_counter",
			Help:           "help",
			Labels:         Labels{"f_": "ok", "f&": "ok"}, // scrubbing introduces duplicate label names
			VariableLabels: []string{"var"},
		})
		assert.Error(t, err, "Expected an error constructing a counter vector with invalid options.")
	})

	t.Run("duplicate variable label names", func(t *testing.T) {
		_, err := r.NewCounterVector(Opts{
			Name:           "test_counter",
			Help:           "help",
			VariableLabels: []string{"var", "var"},
		})
		assert.Error(t, err, "Expected an error constructing a counter vector with invalid options.")
	})
}
