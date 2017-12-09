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
	_, err := scope.NewCounter(spec)
	assert.NoError(t, err, "Failed first registration.")

	t.Run("same type", func(t *testing.T) {
		// You can't reuse specs with the same metric type.
		_, err := scope.NewCounter(spec)
		assert.Error(t, err)
	})

	t.Run("different type", func(t *testing.T) {
		// Even if you change the metric type, you still can't re-use metadata.
		_, err := scope.NewGauge(spec)
		assert.Error(t, err)
		_, err = scope.NewHistogram(HistogramSpec{
			Spec:    spec,
			Unit:    time.Nanosecond,
			Buckets: []int64{1, 2},
		})
		assert.Error(t, err)
	})

	t.Run("different help", func(t *testing.T) {
		// Changing the help string doesn't change the metric's identity.
		_, err := scope.NewCounter(Spec{
			Name: "foo",
			Help: "different help",
		})
		assert.Error(t, err)
	})

	t.Run("added dimensions", func(t *testing.T) {
		// Can't have the same metric name with added dimensions.
		_, err := scope.NewCounter(Spec{
			Name:   "foo",
			Help:   "help",
			Labels: Labels{"bar": "baz"},
		})
		assert.Error(t, err)
	})

	t.Run("different dimensions", func(t *testing.T) {
		// Even if the number of dimensions is the same, metrics with the same
		// name must have the same dimensions.
		_, err := scope.NewCounter(Spec{
			Name:   "dimensions",
			Help:   "help",
			Labels: Labels{"bar": "baz"},
		})
		assert.NoError(t, err, "Failed to register new metric.")
		_, err = scope.NewCounter(Spec{
			Name:   "dimensions",
			Help:   "help",
			Labels: Labels{"bing": "quux"},
		})
		assert.Error(t, err)
	})

	t.Run("same dimensions", func(t *testing.T) {
		// If a metric has the same name and dimensions, the label values may
		// change. This allows users to (inefficiently) create what are
		// effectively vectors - a collection of metrics with the same name and
		// label names, but different label values.
		_, err := scope.NewCounter(Spec{
			Name:   "dimensions",
			Help:   "help",
			Labels: Labels{"bar": "quux"},
		})
		assert.NoError(t, err)
	})

	t.Run("duplicate scrubbed name", func(t *testing.T) {
		// Uniqueness is enforced after the metric name is scrubbed.
		_, err := scope.NewCounter(Spec{
			Name: "scrubbed_name",
			Help: "help",
		})
		assert.NoError(t, err, "Failed to register new metric.")
		_, err = scope.NewCounter(Spec{
			Name: "scrubbed&name",
			Help: "help",
		})
		assert.Error(t, err)
	})

	t.Run("duplicate scrubbed dimensions", func(t *testing.T) {
		// Uniqueness is enforced after labels are scrubbed.
		_, err := scope.NewCounter(Spec{
			Name:   "scrubbed_dimensions",
			Help:   "help",
			Labels: Labels{"b_r": "baz"},
		})
		assert.NoError(t, err, "Failed to register new metric.")
		_, err = scope.NewCounter(Spec{
			Name:   "scrubbed_dimensions",
			Help:   "help",
			Labels: Labels{"b&r": "baz"},
		})
		assert.Error(t, err)
	})

	t.Run("constant label name specified twice", func(t *testing.T) {
		// Within a single user-supplied set of labels, scrubbing may not
		// introduce duplicates.
		_, err = scope.NewCounter(Spec{
			Name:   "user_error_constant_labels",
			Help:   "help",
			Labels: Labels{"b_r": "baz", "b&r": "baz"},
		})
		assert.Error(t, err)
	})
}

func TestVectorMetricDuplicates(t *testing.T) {
	scope := New().Scope()
	spec := Spec{
		Name:           "foo",
		Help:           "help",
		VariableLabels: []string{"foo"},
	}
	_, err := scope.NewCounterVector(spec)
	assert.NoError(t, err, "Failed first registration.")

	t.Run("same type", func(t *testing.T) {
		// You can't reuse specs with the same metric type.
		_, err := scope.NewCounterVector(spec)
		assert.Error(t, err, "Unexpected success re-using vector metrics metadata.")
	})

	t.Run("different type", func(t *testing.T) {
		// Even if you change the metric type, you still can't re-use metadata.
		_, err := scope.NewGaugeVector(spec)
		assert.Error(t, err, "Unexpected success re-using vector metrics metadata.")
		_, err = scope.NewHistogramVector(HistogramSpec{
			Spec:    spec,
			Unit:    time.Nanosecond,
			Buckets: []int64{1, 2},
		})
		assert.Error(t, err, "Unexpected success re-using vector metrics metadata.")
	})

	t.Run("different type and mixed labels", func(t *testing.T) {
		// If we change the type and make some constant labels variable, we still
		// can't re-use metadata.
		_, err := scope.NewCounterVector(Spec{
			Name:           "test_different_type_mixed_labels",
			Help:           "help",
			Labels:         Labels{"foo": "ok"},
			VariableLabels: []string{"bar"},
		})
		require.NoError(t, err, "Failed to create initial metric.")
		_, err = scope.NewGaugeVector(Spec{
			Name:           "test_different_type_mixed_labels",
			Help:           "help",
			VariableLabels: []string{"foo", "bar"},
		})
		require.Error(t, err)
	})

	t.Run("different help", func(t *testing.T) {
		// Changing the help string doesn't change the metric's identity.
		_, err := scope.NewCounterVector(Spec{
			Name:           "foo",
			Help:           "different help",
			VariableLabels: []string{"foo"},
		})
		assert.Error(t, err)
	})

	t.Run("added dimensions", func(t *testing.T) {
		// Can't have the same metric name with added dimensions.
		_, err := scope.NewCounterVector(Spec{
			Name:           "foo",
			Help:           "help",
			VariableLabels: []string{"foo"},
			Labels:         Labels{"bar": "baz"},
		})
		assert.Error(t, err, "Shouldn't be able to add constant labels.")
		_, err = scope.NewCounterVector(Spec{
			Name:           "foo",
			Help:           "help",
			VariableLabels: []string{"foo", "bar"},
		})
		assert.Error(t, err, "Shouldn't be able to add variable labels.")
	})

	t.Run("different dimensions", func(t *testing.T) {
		// Even if the number of dimensions is the same, metrics with the same
		// name must have the same dimensions.
		_, err := scope.NewCounterVector(Spec{
			Name:           "foo",
			Help:           "help",
			VariableLabels: []string{"bar"},
		})
		assert.Error(t, err)
	})

	t.Run("same dimensions", func(t *testing.T) {
		// If a metric has the same name and dimensions, the label values
		// may change. (Again, this would be more efficiently modeled as a
		// higher-dimensionality vector.)
		_, err := scope.NewCounterVector(Spec{
			Name:           "dimensions",
			Help:           "help",
			Labels:         Labels{"bar": "baz"},
			VariableLabels: []string{"foo"},
		})
		assert.NoError(t, err)
		_, err = scope.NewCounterVector(Spec{
			Name:           "dimensions",
			Help:           "help",
			Labels:         Labels{"bar": "quux"},
			VariableLabels: []string{"foo"},
		})
		assert.NoError(t, err)
	})

	t.Run("vectors own dimensions", func(t *testing.T) {
		// If a vector with given dimensions exists, scalars that could be part of
		// that vector may not exist. In other words, for a given set of
		// dimensions, users can't sometimes use a vector and sometimes use a la
		// carte scalars.

		// dims: foo, baz
		_, err := scope.NewCounterVector(Spec{
			Name:           "ownership",
			Help:           "help",
			Labels:         Labels{"foo": "bar"},
			VariableLabels: []string{"baz"},
		})
		require.NoError(t, err)

		// same dims
		_, err = scope.NewCounter(Spec{
			Name:   "ownership",
			Help:   "help",
			Labels: Labels{"foo": "bar", "baz": "quux"},
		})
		require.Error(t, err)
	})

	t.Run("duplicate scrubbed name", func(t *testing.T) {
		// Uniqueness is enforced after the metric name is scrubbed.
		_, err := scope.NewCounterVector(Spec{
			Name:           "scrubbed_name",
			Help:           "help",
			VariableLabels: []string{"bar"},
		})
		assert.NoError(t, err, "Failed to register new metric.")
		_, err = scope.NewCounterVector(Spec{
			Name:           "scrubbed&name",
			Help:           "help",
			VariableLabels: []string{"bar"},
		})
		assert.Error(t, err)
	})

	t.Run("duplicate scrubbed dimensions", func(t *testing.T) {
		// Uniqueness is enforced after labels are scrubbed.
		_, err := scope.NewCounterVector(Spec{
			Name:           "scrubbed_dimensions",
			Help:           "help",
			Labels:         Labels{"b_r": "baz"},
			VariableLabels: []string{"q__x"},
		})
		assert.NoError(t, err, "Failed to register new metric.")
		_, err = scope.NewCounterVector(Spec{
			Name:           "scrubbed_dimensions",
			Help:           "help",
			Labels:         Labels{"b&r": "baz"},
			VariableLabels: []string{"q&&x"},
		})
		assert.Error(t, err)
	})

	t.Run("constant label name specified twice", func(t *testing.T) {
		// Within a single user-supplied set of constant labels, scrubbing may not
		// introduce duplicates.
		_, err = scope.NewCounterVector(Spec{
			Name:           "user_error_constant_labels",
			Help:           "help",
			Labels:         Labels{"b_r": "baz", "b&r": "baz"},
			VariableLabels: []string{"quux"},
		})
		assert.Error(t, err)
	})

	t.Run("variable label name specified twice", func(t *testing.T) {
		// Within a single user-supplied set of variable labels, scrubbing may not
		// introduce duplicates.
		_, err = scope.NewCounterVector(Spec{
			Name:           "user_error_variable_labels",
			Help:           "help",
			VariableLabels: []string{"f__", "f&&"},
		})
		assert.Error(t, err)
	})

	t.Run("constant and variable label name overlap", func(t *testing.T) {
		_, err = scope.NewCounterVector(Spec{
			Name:           "user_error_label_overlaps",
			Help:           "help",
			Labels:         Labels{"foo": "one"},
			VariableLabels: []string{"foo"},
		})
		assert.Error(t, err)
	})
}

func TestLabeledPrecedence(t *testing.T) {
	root := New()
	scope := root.Scope().Labeled(Labels{"foo": "bar"}).Labeled(Labels{"foo": "baz"})
	_, err := scope.NewCounter(Spec{
		Name: "test_counter",
		Help: "help",
	})
	require.NoError(t, err, "Failed to create counter.")
	snap := root.Snapshot()
	require.Equal(t, 1, len(snap.Counters), "Unexpected number of counters.")
	assert.Equal(t, Snapshot{
		Name:   "test_counter",
		Labels: Labels{"foo": "baz"},
	}, snap.Counters[0], "Unexpected counter snapshot.")
}

func TestLabeledAutoScrubbing(t *testing.T) {
	root := New()
	scope := root.Scope().Labeled(Labels{
		"invalid-prometheus-name": "foo",
		"tally":                   "invalid!value",
		"valid":                   "ok",
	})
	vec, err := scope.NewCounterVector(Spec{
		Name:           "test_counter",
		Help:           "help",
		VariableLabels: []string{"invalid_var_name!"},
	})
	vec.MustGet("invalid_var_name!", "ok").Inc()

	require.NoError(t, err, "Failed to create counter.")
	snap := root.Snapshot()
	require.Equal(t, 1, len(snap.Counters), "Unexpected number of counters.")
	assert.Equal(t, Snapshot{
		Name: "test_counter",
		Labels: Labels{
			"invalid_prometheus_name": "foo",
			"tally":                   "invalid_value",
			"valid":                   "ok",
			"invalid_var_name_":       "ok",
		},
		Value: 1,
	}, snap.Counters[0], "Unexpected counter snapshot.")
}

func TestLabelScrubbingUniqueness(t *testing.T) {
	t.Run("duplicate const names", func(t *testing.T) {
		scope := New().Scope()
		_, err := scope.NewCounter(Spec{
			Name: "test",
			Help: "help",
			Labels: Labels{
				"foo_bar": "baz",
				"foo!bar": "baz",
			},
		})
		assert.Error(t, err, "Expected error when scrubbing duplicates label names.")
	})
	t.Run("duplicate variable names", func(t *testing.T) {
		scope := New().Scope()
		_, err := scope.NewCounterVector(Spec{
			Name:           "test",
			Help:           "help",
			VariableLabels: []string{"foo_bar", "foo!bar"},
		})
		assert.Error(t, err, "Expected error when scrubbing duplicates label names.")
	})
}
