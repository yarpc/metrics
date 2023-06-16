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

	promproto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"
	bucketpkg "go.uber.org/net/metrics/bucket"
	"go.uber.org/net/metrics/tallypush"
)

func uint64ptr(i uint64) *uint64 {
	return &i
}

func float64ptr(i float64) *float64 {
	return &i
}

func TestHistogram(t *testing.T) {
	root := New()
	s := root.Scope().Tagged(Tags{"service": "users"})

	t.Run("duplicate constant tag names", func(t *testing.T) {
		_, err := s.Histogram(HistogramSpec{
			Spec: Spec{
				Name:      "test_latency_ns",
				Help:      "Some help.",
				ConstTags: Tags{"f_": "ok", "f&": "ok"}, // scrubbing introduces duplicate names
			},
			Unit:    time.Nanosecond,
			Buckets: []int64{10, 50, 100},
		})
		assert.Error(t, err, "Expected an error constructing a histogram with invalid spec.")
	})

	t.Run("valid spec", func(t *testing.T) {
		h, err := s.Histogram(HistogramSpec{
			Spec: Spec{
				Name:      "test_latency_ns",
				Help:      "Some help.",
				ConstTags: Tags{"foo": "bar"},
			},
			Unit:    time.Nanosecond,
			Buckets: []int64{10, 50, 100},
		})
		require.NoError(t, err, "Unexpected construction error.")

		h.Observe(-1)
		h.IncBucket(0)
		h.Observe(10)
		h.IncBucket(75)
		h.Observe(150)

		snap := root.Snapshot()
		require.Equal(t, 1, len(snap.Histograms), "Unexpected number of histogram snapshots.")
		got := snap.Histograms[0]
		assert.Equal(t, HistogramSnapshot{
			Unit:   time.Nanosecond,
			Name:   "test_latency_ns",
			Tags:   Tags{"foo": "bar", "service": "users"},
			Values: []int64{10, 10, 10, 100, math.MaxInt64},
		}, got, "Unexpected histogram snapshot.")
	})

	t.Run("prometheus export", func(t *testing.T) {
		h, err := s.Histogram(HistogramSpec{
			Spec: Spec{
				Name: "test_histogram",
				Help: "Some help.",
			},
			Unit:    time.Nanosecond,
			Buckets: []int64{10, 50, 100},
		})
		require.NoError(t, err, "Unexpected construction error.")

		h.IncBucket(5)
		h.IncBucket(45)
		h.IncBucket(90)

		expectedHistogram := &promproto.Histogram{
			SampleCount: uint64ptr(3),
			SampleSum:   float64ptr(140),
			Bucket: []*promproto.Bucket{
				{
					CumulativeCount: uint64ptr(1),
					UpperBound:      float64ptr(10),
				},
				{
					CumulativeCount: uint64ptr(2),
					UpperBound:      float64ptr(50),
				},
				{
					CumulativeCount: uint64ptr(3),
					UpperBound:      float64ptr(100),
				},
			},
		}
		assert.Equal(t, expectedHistogram, h.metric().Histogram)
	})
}

func TestHistogramVector(t *testing.T) {
	tests := []struct {
		desc string
		spec HistogramSpec
		f    func(testing.TB, *Scope, HistogramSpec)
		want HistogramSnapshot
	}{
		{
			desc: "valid tags",
			spec: HistogramSpec{
				Spec: Spec{
					Name:    "test_latency_ms",
					Help:    "Some help.",
					VarTags: []string{"var"},
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			f: func(t testing.TB, s *Scope, spec HistogramSpec) {
				vec, err := s.HistogramVector(spec)
				require.NoError(t, err, "Unexpected error constructing vector.")
				h, err := vec.Get("var", "x")
				require.NoError(t, err, "Unexpected error getting a counter with correct number of tags.")
				h.Observe(time.Millisecond)
				vec.MustGet("var", "x").Observe(time.Millisecond)
			},
			want: HistogramSnapshot{
				Name:   "test_latency_ms",
				Tags:   Tags{"var": "x"},
				Unit:   time.Millisecond,
				Values: []int64{1000, 1000},
			},
		},
		{
			desc: "invalid tags",
			spec: HistogramSpec{
				Spec: Spec{
					Name:    "test_latency_ms",
					Help:    "Some help.",
					VarTags: []string{"var"},
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			f: func(t testing.TB, s *Scope, spec HistogramSpec) {
				vec, err := s.HistogramVector(spec)
				require.NoError(t, err, "Unexpected error constructing vector.")
				h, err := vec.Get("var", "x!")
				require.NoError(t, err, "Unexpected error getting a counter with correct number of tags.")
				h.Observe(time.Millisecond)
				vec.MustGet("var", "x!").Observe(time.Millisecond)
			},
			want: HistogramSnapshot{
				Name:   "test_latency_ms",
				Tags:   Tags{"var": "x_"},
				Unit:   time.Millisecond,
				Values: []int64{1000, 1000},
			},
		},
		{
			desc: "wrong number of tag values",
			spec: HistogramSpec{
				Spec: Spec{
					Name:    "test_latency_ms",
					Help:    "Some help.",
					VarTags: []string{"var"},
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			f: func(t testing.TB, s *Scope, spec HistogramSpec) {
				vec, err := s.HistogramVector(spec)
				require.NoError(t, err, "Unexpected error constructing vector.")
				_, err = vec.Get("var", "x", "var2", "y")
				require.Error(t, err, "Unexpected success calling Get with incorrect number of tags.")
				require.Panics(
					t,
					func() { vec.MustGet("var", "x", "var2", "y") },
					"Expected a panic using MustGet with the wrong number of tags.",
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			root := New()
			tt.f(t, root.Scope(), tt.spec)
			snap := root.Snapshot()
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
	root := New()
	spec := HistogramSpec{
		Spec: Spec{
			Name:    "test_latency_ms",
			Help:    "Some help.",
			VarTags: []string{"var"},
		},
		Unit:    time.Millisecond,
		Buckets: []int64{1000},
	}
	vec, err := root.Scope().HistogramVector(spec)
	require.NoError(t, err, "Unexpected error constructing vector.")

	x, err := vec.Get("var", "x")
	require.NoError(t, err, "Unexpected error calling Get.")

	y, err := vec.Get("var", "y")
	require.NoError(t, err, "Unexpected error calling Get.")

	x.Observe(time.Millisecond)
	y.Observe(time.Millisecond)

	snap := root.Snapshot()
	require.Equal(t, 2, len(snap.Histograms), "Unexpected number of histogram snapshots.")

	assert.Equal(t, HistogramSnapshot{
		Name:   "test_latency_ms",
		Tags:   Tags{"var": "x"},
		Unit:   time.Millisecond,
		Values: []int64{1000},
	}, snap.Histograms[0], "Unexpected first histogram snapshot.")
	assert.Equal(t, HistogramSnapshot{
		Name:   "test_latency_ms",
		Tags:   Tags{"var": "y"},
		Unit:   time.Millisecond,
		Values: []int64{1000},
	}, snap.Histograms[1], "Unexpected second histogram snapshot.")
}

func BenchmarkHistogram(b *testing.B) {
	pusher := tallypush.New(tally.NoopScope)
	name := ""
	hist := newHistogram(metadata{
		Name: &name,
	}, time.Millisecond, bucketpkg.NewRPCLatency())
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hist.push(pusher)
	}
}
