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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScalarMetricDuplicates(t *testing.T) {
	scope := New().Scope()
	spec := Spec{
		Name: "foo",
		Help: "help",
	}
	_, err := scope.Counter(spec)
	assert.NoError(t, err, "Failed first registration.")

	t.Run("same type", func(t *testing.T) {
		// You can't reuse specs with the same metric type.
		_, err := scope.Counter(spec)
		assert.Error(t, err)
	})

	t.Run("different type", func(t *testing.T) {
		// Even if you change the metric type, you still can't re-use metadata.
		_, err := scope.Gauge(spec)
		assert.Error(t, err)
		_, err = scope.Histogram(HistogramSpec{
			Spec:    spec,
			Unit:    time.Nanosecond,
			Buckets: []int64{1, 2},
		})
		assert.Error(t, err)
	})

	t.Run("different help", func(t *testing.T) {
		// Changing the help string doesn't change the metric's identity.
		_, err := scope.Counter(Spec{
			Name: "foo",
			Help: "different help",
		})
		assert.Error(t, err)
	})

	t.Run("added dimensions", func(t *testing.T) {
		// Can't have the same metric name with added dimensions.
		_, err := scope.Counter(Spec{
			Name:      "foo",
			Help:      "help",
			ConstTags: Tags{"bar": "baz"},
		})
		assert.Error(t, err)
	})

	t.Run("different dimensions", func(t *testing.T) {
		// Even if the number of dimensions is the same, metrics with the same
		// name must have the same dimensions.
		_, err := scope.Counter(Spec{
			Name:      "dimensions",
			Help:      "help",
			ConstTags: Tags{"bar": "baz"},
		})
		assert.NoError(t, err, "Failed to register new metric.")
		_, err = scope.Counter(Spec{
			Name:      "dimensions",
			Help:      "help",
			ConstTags: Tags{"bing": "quux"},
		})
		assert.Error(t, err)
	})

	t.Run("same dimensions", func(t *testing.T) {
		// If a metric has the same name and dimensions, the tag values may
		// change. This allows users to (inefficiently) create what are
		// effectively vectors - a collection of metrics with the same name and
		// tag names, but different tag values.
		_, err := scope.Counter(Spec{
			Name:      "dimensions",
			Help:      "help",
			ConstTags: Tags{"bar": "quux"},
		})
		assert.NoError(t, err)
	})

	t.Run("duplicate scrubbed name", func(t *testing.T) {
		// Uniqueness is enforced after the metric name is scrubbed.
		_, err := scope.Counter(Spec{
			Name: "scrubbed_name",
			Help: "help",
		})
		assert.NoError(t, err, "Failed to register new metric.")
		_, err = scope.Counter(Spec{
			Name: "scrubbed&name",
			Help: "help",
		})
		assert.Error(t, err)
	})

	t.Run("duplicate scrubbed dimensions", func(t *testing.T) {
		// Uniqueness is enforced after tags are scrubbed.
		_, err := scope.Counter(Spec{
			Name:      "scrubbed_dimensions",
			Help:      "help",
			ConstTags: Tags{"b_r": "baz"},
		})
		assert.NoError(t, err, "Failed to register new metric.")
		_, err = scope.Counter(Spec{
			Name:      "scrubbed_dimensions",
			Help:      "help",
			ConstTags: Tags{"b&r": "baz"},
		})
		assert.Error(t, err)
	})

	t.Run("constant tag name specified twice", func(t *testing.T) {
		// Within a single user-supplied set of tags, scrubbing may not
		// introduce duplicates.
		_, err = scope.Counter(Spec{
			Name:      "user_error_constant_tags",
			Help:      "help",
			ConstTags: Tags{"b_r": "baz", "b&r": "baz"},
		})
		assert.Error(t, err)
	})
}

func TestVectorMetricDuplicates(t *testing.T) {
	scope := New().Scope()
	spec := Spec{
		Name:    "foo",
		Help:    "help",
		VarTags: []string{"foo"},
	}
	_, err := scope.CounterVector(spec)
	assert.NoError(t, err, "Failed first registration.")

	t.Run("same type", func(t *testing.T) {
		// You can't reuse specs with the same metric type.
		_, err := scope.CounterVector(spec)
		assert.Error(t, err, "Unexpected success re-using vector metrics metadata.")
	})

	t.Run("different type", func(t *testing.T) {
		// Even if you change the metric type, you still can't re-use metadata.
		_, err := scope.GaugeVector(spec)
		assert.Error(t, err, "Unexpected success re-using vector metrics metadata.")
		_, err = scope.HistogramVector(HistogramSpec{
			Spec:    spec,
			Unit:    time.Nanosecond,
			Buckets: []int64{1, 2},
		})
		assert.Error(t, err, "Unexpected success re-using vector metrics metadata.")
	})

	t.Run("different type and mixed tags", func(t *testing.T) {
		// If we change the type and make some constant tags variable, we still
		// can't re-use metadata.
		_, err := scope.CounterVector(Spec{
			Name:      "test_different_type_mixed_tags",
			Help:      "help",
			ConstTags: Tags{"foo": "ok"},
			VarTags:   []string{"bar"},
		})
		require.NoError(t, err, "Failed to create initial metric.")
		_, err = scope.GaugeVector(Spec{
			Name:    "test_different_type_mixed_tags",
			Help:    "help",
			VarTags: []string{"foo", "bar"},
		})
		require.Error(t, err)
	})

	t.Run("different help", func(t *testing.T) {
		// Changing the help string doesn't change the metric's identity.
		_, err := scope.CounterVector(Spec{
			Name:    "foo",
			Help:    "different help",
			VarTags: []string{"foo"},
		})
		assert.Error(t, err)
	})

	t.Run("added dimensions", func(t *testing.T) {
		// Can't have the same metric name with added dimensions.
		_, err := scope.CounterVector(Spec{
			Name:      "foo",
			Help:      "help",
			VarTags:   []string{"foo"},
			ConstTags: Tags{"bar": "baz"},
		})
		assert.Error(t, err, "Shouldn't be able to add constant tags.")
		_, err = scope.CounterVector(Spec{
			Name:    "foo",
			Help:    "help",
			VarTags: []string{"foo", "bar"},
		})
		assert.Error(t, err, "Shouldn't be able to add variable tags.")
	})

	t.Run("different dimensions", func(t *testing.T) {
		// Even if the number of dimensions is the same, metrics with the same
		// name must have the same dimensions.
		_, err := scope.CounterVector(Spec{
			Name:    "foo",
			Help:    "help",
			VarTags: []string{"bar"},
		})
		assert.Error(t, err)
	})

	t.Run("same dimensions", func(t *testing.T) {
		// If a metric has the same name and dimensions, the tag values
		// may change. (Again, this would be more efficiently modeled as a
		// higher-dimensionality vector.)
		_, err := scope.CounterVector(Spec{
			Name:      "dimensions",
			Help:      "help",
			ConstTags: Tags{"bar": "baz"},
			VarTags:   []string{"foo"},
		})
		assert.NoError(t, err)
		_, err = scope.CounterVector(Spec{
			Name:      "dimensions",
			Help:      "help",
			ConstTags: Tags{"bar": "quux"},
			VarTags:   []string{"foo"},
		})
		assert.NoError(t, err)
	})

	t.Run("vectors own dimensions", func(t *testing.T) {
		// If a vector with given dimensions exists, scalars that could be part of
		// that vector may not exist. In other words, for a given set of
		// dimensions, users can't sometimes use a vector and sometimes use a la
		// carte scalars.

		// dims: foo, baz
		_, err := scope.CounterVector(Spec{
			Name:      "ownership",
			Help:      "help",
			ConstTags: Tags{"foo": "bar"},
			VarTags:   []string{"baz"},
		})
		require.NoError(t, err)

		// same dims
		_, err = scope.Counter(Spec{
			Name:      "ownership",
			Help:      "help",
			ConstTags: Tags{"foo": "bar", "baz": "quux"},
		})
		require.Error(t, err)
	})

	t.Run("duplicate scrubbed name", func(t *testing.T) {
		// Uniqueness is enforced after the metric name is scrubbed.
		_, err := scope.CounterVector(Spec{
			Name:    "scrubbed_name",
			Help:    "help",
			VarTags: []string{"bar"},
		})
		assert.NoError(t, err, "Failed to register new metric.")
		_, err = scope.CounterVector(Spec{
			Name:    "scrubbed&name",
			Help:    "help",
			VarTags: []string{"bar"},
		})
		assert.Error(t, err)
	})

	t.Run("duplicate scrubbed dimensions", func(t *testing.T) {
		// Uniqueness is enforced after tags are scrubbed.
		_, err := scope.CounterVector(Spec{
			Name:      "scrubbed_dimensions",
			Help:      "help",
			ConstTags: Tags{"b_r": "baz"},
			VarTags:   []string{"q__x"},
		})
		assert.NoError(t, err, "Failed to register new metric.")
		_, err = scope.CounterVector(Spec{
			Name:      "scrubbed_dimensions",
			Help:      "help",
			ConstTags: Tags{"b&r": "baz"},
			VarTags:   []string{"q&&x"},
		})
		assert.Error(t, err)
	})

	t.Run("constant tag name specified twice", func(t *testing.T) {
		// Within a single user-supplied set of constant tags, scrubbing may not
		// introduce duplicates.
		_, err = scope.CounterVector(Spec{
			Name:      "user_error_constant_tags",
			Help:      "help",
			ConstTags: Tags{"b_r": "baz", "b&r": "baz"},
			VarTags:   []string{"quux"},
		})
		assert.Error(t, err)
	})

	t.Run("variable tag name specified twice", func(t *testing.T) {
		// Within a single user-supplied set of variable tags, scrubbing may not
		// introduce duplicates.
		_, err = scope.CounterVector(Spec{
			Name:    "user_error_variable_tags",
			Help:    "help",
			VarTags: []string{"f__", "f&&"},
		})
		assert.Error(t, err)
	})

	t.Run("constant and variable tag name overlap", func(t *testing.T) {
		_, err = scope.CounterVector(Spec{
			Name:      "user_error_tag_overlaps",
			Help:      "help",
			ConstTags: Tags{"foo": "one"},
			VarTags:   []string{"foo"},
		})
		assert.Error(t, err)
	})
}

func TestTaggedPrecedence(t *testing.T) {
	root := New()
	scope := root.Scope().Tagged(Tags{"foo": "bar"}).Tagged(Tags{"foo": "baz"})
	_, err := scope.Counter(Spec{
		Name: "test_counter",
		Help: "help",
	})
	require.NoError(t, err, "Failed to create counter.")
	snap := root.Snapshot()
	require.Equal(t, 1, len(snap.Counters), "Unexpected number of counters.")
	assert.Equal(t, Snapshot{
		Name: "test_counter",
		Tags: Tags{"foo": "baz"},
	}, snap.Counters[0], "Unexpected counter snapshot.")
}

func TestTaggedAutoScrubbing(t *testing.T) {
	root := New()
	scope := root.Scope().Tagged(Tags{
		"invalid-prometheus-name": "foo",
		"tally":                   "invalid!value",
		"valid":                   "ok",
	})
	vec, err := scope.CounterVector(Spec{
		Name:    "test_counter",
		Help:    "help",
		VarTags: []string{"invalid_var_name!"},
	})
	vec.MustGet("invalid_var_name!", "ok").Inc()

	require.NoError(t, err, "Failed to create counter.")
	snap := root.Snapshot()
	require.Equal(t, 1, len(snap.Counters), "Unexpected number of counters.")
	assert.Equal(t, Snapshot{
		Name: "test_counter",
		Tags: Tags{
			"invalid_prometheus_name": "foo",
			"tally":                   "invalid_value",
			"valid":                   "ok",
			"invalid_var_name_":       "ok",
		},
		Value: 1,
	}, snap.Counters[0], "Unexpected counter snapshot.")
}

func TestTagScrubbingUniqueness(t *testing.T) {
	t.Run("duplicate const names", func(t *testing.T) {
		scope := New().Scope()
		_, err := scope.Counter(Spec{
			Name: "test",
			Help: "help",
			ConstTags: Tags{
				"foo_bar": "baz",
				"foo!bar": "baz",
			},
		})
		assert.Error(t, err, "Expected error when scrubbing duplicates tag names.")
	})
	t.Run("duplicate variable names", func(t *testing.T) {
		scope := New().Scope()
		_, err := scope.CounterVector(Spec{
			Name:    "test",
			Help:    "help",
			VarTags: []string{"foo_bar", "foo!bar"},
		})
		assert.Error(t, err, "Expected error when scrubbing duplicates tag names.")
	})
}
