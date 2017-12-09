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
)

func TestSpecValidation(t *testing.T) {
	tests := []struct {
		desc     string
		spec     Spec
		scalarOK bool
		vecOK    bool
	}{
		{
			desc: "valid names",
			spec: Spec{
				Name: "fOo123",
				Help: "Some help.",
			},
			scalarOK: true,
			vecOK:    false,
		},
		{
			desc: "valid names & constant tags",
			spec: Spec{
				Name:      "foo",
				Help:      "Some help.",
				ConstTags: Tags{"foo": "bar"},
			},
			scalarOK: true,
			vecOK:    false,
		},
		{
			desc: "name with Tally-forbidden characters",
			spec: Spec{
				Name: "foo:bar",
				Help: "Some help.",
			},
			scalarOK: true,
			vecOK:    false,
		},
		{
			desc: "no name",
			spec: Spec{
				Help: "Some help.",
			},
			scalarOK: false,
			vecOK:    false,
		},
		{
			desc: "no help",
			spec: Spec{
				Name: "foo",
			},
			scalarOK: false,
			vecOK:    false,
		},
		{
			desc: "valid names but invalid tag key",
			spec: Spec{
				Name:      "foo",
				Help:      "Some help.",
				ConstTags: Tags{"foo:foo": "bar"},
			},
			scalarOK: true,
			vecOK:    false,
		},
		{
			desc: "valid names but invalid tag value",
			spec: Spec{
				Name:      "foo",
				Help:      "Some help.",
				ConstTags: Tags{"foo": "bar:bar"},
			},
			scalarOK: true,
			vecOK:    false,
		},
		{
			desc: "valid names & variable tags",
			spec: Spec{
				Name:    "foo",
				Help:    "Some help.",
				VarTags: []string{"baz"},
			},
			scalarOK: false,
			vecOK:    true,
		},
		{
			desc: "valid names, constant tags, & variable tags",
			spec: Spec{
				Name:      "foo",
				Help:      "Some help.",
				ConstTags: Tags{"foo": "bar"},
				VarTags:   []string{"baz"},
			},
			scalarOK: false,
			vecOK:    true,
		},
		{
			desc: "valid names & constant tags, but invalid variable tags",
			spec: Spec{
				Name:      "foo",
				Help:      "Some help.",
				ConstTags: Tags{"foo": "bar"},
				VarTags:   []string{"baz:baz"},
			},
			scalarOK: false,
			vecOK:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if tt.scalarOK {
				assertScalarSpecOK(t, tt.spec)
			} else {
				assertScalarSpecFail(t, tt.spec)
			}
			if tt.vecOK {
				assertVectorSpecOK(t, tt.spec)
			} else {
				assertVectorSpecFail(t, tt.spec)
			}
		})
	}
}

func TestHistogramSpecValidation(t *testing.T) {
	tests := []struct {
		desc     string
		spec     HistogramSpec
		scalarOK bool
		vecOK    bool
	}{
		{
			desc: "valid names",
			spec: HistogramSpec{
				Spec: Spec{
					Name: "fOo123",
					Help: "Some help.",
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: true,
			vecOK:    false,
		},
		{
			desc: "valid names & constant tags",
			spec: HistogramSpec{
				Spec: Spec{
					Name:      "foo",
					Help:      "Some help.",
					ConstTags: Tags{"foo": "bar"},
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: true,
			vecOK:    false,
		},
		{
			desc: "name with Tally-forbidden characters",
			spec: HistogramSpec{
				Spec: Spec{
					Name: "foo:bar",
					Help: "Some help.",
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: true,
			vecOK:    false,
		},
		{
			desc: "no name",
			spec: HistogramSpec{
				Spec: Spec{
					Help: "Some help.",
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: false,
			vecOK:    false,
		},
		{
			desc: "no help",
			spec: HistogramSpec{
				Spec: Spec{
					Name: "foo",
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: false,
			vecOK:    false,
		},
		{
			desc: "valid names but invalid tag key",
			spec: HistogramSpec{
				Spec: Spec{
					Name:      "foo",
					Help:      "Some help.",
					ConstTags: Tags{"foo:foo": "bar"},
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: true,
			vecOK:    false,
		},
		{
			desc: "valid names but invalid tag value",
			spec: HistogramSpec{
				Spec: Spec{
					Name:      "foo",
					Help:      "Some help.",
					ConstTags: Tags{"foo": "bar:bar"},
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: true,
			vecOK:    false,
		},
		{
			desc: "valid names & variable tags",
			spec: HistogramSpec{
				Spec: Spec{
					Name:    "foo",
					Help:    "Some help.",
					VarTags: []string{"baz"},
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: false,
			vecOK:    true,
		},
		{
			desc: "valid names, constant tags, & variable tags",
			spec: HistogramSpec{
				Spec: Spec{
					Name:      "foo",
					Help:      "Some help.",
					ConstTags: Tags{"foo": "bar"},
					VarTags:   []string{"baz"},
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: false,
			vecOK:    true,
		},
		{
			desc: "valid names & constant tags, but invalid variable tags",
			spec: HistogramSpec{
				Spec: Spec{
					Name:      "foo",
					Help:      "Some help.",
					ConstTags: Tags{"foo": "bar"},
					VarTags:   []string{"baz:baz"},
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: false,
			vecOK:    true,
		},
		{
			desc: "valid tags, no unit",
			spec: HistogramSpec{
				Spec: Spec{
					Name:      "foo",
					Help:      "Some help.",
					ConstTags: Tags{"foo": "bar"},
					VarTags:   []string{"baz"},
				},
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: false,
			vecOK:    false,
		},
		{
			desc: "valid tags, negative unit",
			spec: HistogramSpec{
				Spec: Spec{
					Name:      "foo",
					Help:      "Some help.",
					ConstTags: Tags{"foo": "bar"},
					VarTags:   []string{"baz"},
				},
				Unit:    -1 * time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: false,
			vecOK:    false,
		},
		{
			desc: "valid tags, no buckets",
			spec: HistogramSpec{
				Spec: Spec{
					Name:      "foo",
					Help:      "Some help.",
					ConstTags: Tags{"foo": "bar"},
					VarTags:   []string{"baz"},
				},
				Unit: time.Millisecond,
			},
			scalarOK: false,
			vecOK:    false,
		},
		{
			desc: "valid tags, buckets out of order",
			spec: HistogramSpec{
				Spec: Spec{
					Name:      "foo",
					Help:      "Some help.",
					ConstTags: Tags{"foo": "bar"},
					VarTags:   []string{"baz"},
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000 * 60, 1000},
			},
			scalarOK: false,
			vecOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if tt.scalarOK {
				assertScalarHistogramSpecOK(t, tt.spec)
			} else {
				assertSimpleHistogramSpecFail(t, tt.spec)
			}
			if tt.vecOK {
				assertVectorHistogramSpecOK(t, tt.spec)
			} else {
				assertVectorHistogramSpecFail(t, tt.spec)
			}
		})
	}
}

func assertScalarSpecOK(t testing.TB, spec Spec) {
	_, err := New().Scope().NewCounter(spec)
	assert.NoError(t, err, "Expected success from NewCounter.")

	_, err = New().Scope().NewGauge(spec)
	assert.NoError(t, err, "Expected success from NewGauge.")
}

func assertScalarSpecFail(t testing.TB, spec Spec) {
	_, err := New().Scope().NewCounter(spec)
	assert.Error(t, err, "Expected an error from NewCounter.")

	_, err = New().Scope().NewGauge(spec)
	assert.Error(t, err, "Expected an error from NewGauge.")
}

func assertVectorSpecOK(t testing.TB, spec Spec) {
	_, err := New().Scope().NewCounterVector(spec)
	assert.NoError(t, err, "Expected success from NewCounterVector.")

	_, err = New().Scope().NewGaugeVector(spec)
	assert.NoError(t, err, "Expected success from NewGaugeVector.")
}

func assertVectorSpecFail(t testing.TB, spec Spec) {
	_, err := New().Scope().NewCounterVector(spec)
	assert.Error(t, err, "Expected an error from NewCounterVector.")

	_, err = New().Scope().NewGaugeVector(spec)
	assert.Error(t, err, "Expected an error from NewGaugeVector.")
}

func assertScalarHistogramSpecOK(t testing.TB, spec HistogramSpec) {
	_, err := New().Scope().NewHistogram(spec)
	assert.NoError(t, err, "Expected success from NewLatencies.")
}

func assertSimpleHistogramSpecFail(t testing.TB, spec HistogramSpec) {
	_, err := New().Scope().NewHistogram(spec)
	assert.Error(t, err, "Expected an error from NewLatencies.")
}

func assertVectorHistogramSpecOK(t testing.TB, spec HistogramSpec) {
	_, err := New().Scope().NewHistogramVector(spec)
	assert.NoError(t, err, "Expected success from NewLatenciesVector.")
}

func assertVectorHistogramSpecFail(t testing.TB, spec HistogramSpec) {
	_, err := New().Scope().NewHistogramVector(spec)
	assert.Error(t, err, "Expected an error from NewLatenciesVector.")
}
