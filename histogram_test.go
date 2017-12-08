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
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistogram(t *testing.T) {
	r, c := New()
	r = r.Labeled(Labels{"service": "users"})

	t.Run("duplicate constant label names", func(t *testing.T) {
		_, err := r.NewHistogram(HistogramSpec{
			Spec: Spec{
				Name:   "test_latency_ns",
				Help:   "Some help.",
				Labels: Labels{"f_": "ok", "f&": "ok"}, // scrubbing introduces duplicate names
			},
			Unit:    time.Nanosecond,
			Buckets: []int64{10, 50, 100},
		})
		assert.Error(t, err, "Expected an error constructing a histogram with invalid spec.")
	})

	t.Run("valid spec", func(t *testing.T) {
		h, err := r.NewHistogram(HistogramSpec{
			Spec: Spec{
				Name:   "test_latency_ns",
				Help:   "Some help.",
				Labels: Labels{"foo": "bar"},
			},
			Unit:    time.Nanosecond,
			Buckets: []int64{10, 50, 100},
		})
		require.NoError(t, err, "Unexpected construction error.")

		h.Observe(-1)
		h.ObserveInt(0)
		h.Observe(10)
		h.ObserveInt(75)
		h.Observe(150)

		snap := c.Snapshot()
		require.Equal(t, 1, len(snap.Histograms), "Unexpected number of histogram snapshots.")
		got := snap.Histograms[0]
		assert.Equal(t, HistogramSnapshot{
			Unit:   time.Nanosecond,
			Name:   "test_latency_ns",
			Labels: Labels{"foo": "bar", "service": "users"},
			Values: []int64{10, 10, 10, 100, math.MaxInt64},
		}, got, "Unexpected histogram snapshot.")
	})
}

func TestHistogramVector(t *testing.T) {
	tests := []struct {
		desc string
		spec HistogramSpec
		f    func(testing.TB, *Registry, HistogramSpec)
		want HistogramSnapshot
	}{
		{
			desc: "valid labels",
			spec: HistogramSpec{
				Spec: Spec{
					Name:           "test_latency_ms",
					Help:           "Some help.",
					VariableLabels: []string{"var"},
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			f: func(t testing.TB, r *Registry, spec HistogramSpec) {
				vec, err := r.NewHistogramVector(spec)
				require.NoError(t, err, "Unexpected error constructing vector.")
				h, err := vec.Get("var", "x")
				require.NoError(t, err, "Unexpected error getting a counter with correct number of labels.")
				h.Observe(time.Millisecond)
				vec.MustGet("var", "x").Observe(time.Millisecond)
			},
			want: HistogramSnapshot{
				Name:   "test_latency_ms",
				Labels: Labels{"var": "x"},
				Unit:   time.Millisecond,
				Values: []int64{1000, 1000},
			},
		},
		{
			desc: "invalid labels",
			spec: HistogramSpec{
				Spec: Spec{
					Name:           "test_latency_ms",
					Help:           "Some help.",
					VariableLabels: []string{"var"},
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			f: func(t testing.TB, r *Registry, spec HistogramSpec) {
				vec, err := r.NewHistogramVector(spec)
				require.NoError(t, err, "Unexpected error constructing vector.")
				h, err := vec.Get("var", "x!")
				require.NoError(t, err, "Unexpected error getting a counter with correct number of labels.")
				h.Observe(time.Millisecond)
				vec.MustGet("var", "x!").Observe(time.Millisecond)
			},
			want: HistogramSnapshot{
				Name:   "test_latency_ms",
				Labels: Labels{"var": "x_"},
				Unit:   time.Millisecond,
				Values: []int64{1000, 1000},
			},
		},
		{
			desc: "wrong number of label values",
			spec: HistogramSpec{
				Spec: Spec{
					Name:           "test_latency_ms",
					Help:           "Some help.",
					VariableLabels: []string{"var"},
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			f: func(t testing.TB, r *Registry, spec HistogramSpec) {
				vec, err := r.NewHistogramVector(spec)
				require.NoError(t, err, "Unexpected error constructing vector.")
				_, err = vec.Get("var", "x", "var2", "y")
				require.Error(t, err, "Unexpected success calling Get with incorrect number of labels.")
				require.Panics(
					t,
					func() { vec.MustGet("var", "x", "var2", "y") },
					"Expected a panic using MustGet with the wrong number of labels.",
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			r, c := New()
			tt.f(t, r, tt.spec)
			snap := c.Snapshot()
			if tt.want.Name != "" {
				require.Equal(t, 1, len(snap.Histograms), "Unexpected number of histogram snapshots.")
				assert.Equal(t, tt.want, snap.Histograms[0], "Unexpected histogram snapshot.")
			} else {
				require.Equal(t, 0, len(snap.Histograms), "Expected no histogram snapshots.")
			}
		})
	}
}

func TestHistogramVectorIndependence(t *testing.T) {
	// Ensure that we're not erroneously sharing state across histograms in a
	// vector.
	r, c := New()
	spec := HistogramSpec{
		Spec: Spec{
			Name:           "test_latency_ms",
			Help:           "Some help.",
			VariableLabels: []string{"var"},
		},
		Unit:    time.Millisecond,
		Buckets: []int64{1000},
	}
	vec, err := r.NewHistogramVector(spec)
	require.NoError(t, err, "Unexpected error constructing vector.")

	x, err := vec.Get("var", "x")
	require.NoError(t, err, "Unexpected error calling Get.")

	y, err := vec.Get("var", "y")
	require.NoError(t, err, "Unexpected error calling Get.")

	x.Observe(time.Millisecond)
	y.Observe(time.Millisecond)

	snap := c.Snapshot()
	require.Equal(t, 2, len(snap.Histograms), "Unexpected number of histogram snapshots.")

	assert.Equal(t, HistogramSnapshot{
		Name:   "test_latency_ms",
		Labels: Labels{"var": "x"},
		Unit:   time.Millisecond,
		Values: []int64{1000},
	}, snap.Histograms[0], "Unexpected first histogram snapshot.")
	assert.Equal(t, HistogramSnapshot{
		Name:   "test_latency_ms",
		Labels: Labels{"var": "y"},
		Unit:   time.Millisecond,
		Values: []int64{1000},
	}, snap.Histograms[1], "Unexpected second histogram snapshot.")
}
