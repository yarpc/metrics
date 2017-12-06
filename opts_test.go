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

func TestOptsValidation(t *testing.T) {
	tests := []struct {
		desc     string
		opts     Opts
		scalarOK bool
		vecOK    bool
	}{
		{
			desc: "valid names",
			opts: Opts{
				Name: "fOo123",
				Help: "Some help.",
			},
			scalarOK: true,
			vecOK:    false,
		},
		{
			desc: "valid names & constant labels",
			opts: Opts{
				Name:   "foo",
				Help:   "Some help.",
				Labels: Labels{"foo": "bar"},
			},
			scalarOK: true,
			vecOK:    false,
		},
		{
			desc: "name with Tally-forbidden characters",
			opts: Opts{
				Name: "foo:bar",
				Help: "Some help.",
			},
			scalarOK: true,
			vecOK:    false,
		},
		{
			desc: "no name",
			opts: Opts{
				Help: "Some help.",
			},
			scalarOK: false,
			vecOK:    false,
		},
		{
			desc: "no help",
			opts: Opts{
				Name: "foo",
			},
			scalarOK: false,
			vecOK:    false,
		},
		{
			desc: "valid names but invalid label key",
			opts: Opts{
				Name:   "foo",
				Help:   "Some help.",
				Labels: Labels{"foo:foo": "bar"},
			},
			scalarOK: true,
			vecOK:    false,
		},
		{
			desc: "valid names but invalid label value",
			opts: Opts{
				Name:   "foo",
				Help:   "Some help.",
				Labels: Labels{"foo": "bar:bar"},
			},
			scalarOK: true,
			vecOK:    false,
		},
		{
			desc: "valid names & variable labels",
			opts: Opts{
				Name:           "foo",
				Help:           "Some help.",
				VariableLabels: []string{"baz"},
			},
			scalarOK: false,
			vecOK:    true,
		},
		{
			desc: "valid names, constant labels, & variable labels",
			opts: Opts{
				Name:           "foo",
				Help:           "Some help.",
				Labels:         Labels{"foo": "bar"},
				VariableLabels: []string{"baz"},
			},
			scalarOK: false,
			vecOK:    true,
		},
		{
			desc: "valid names & constant labels, but invalid variable labels",
			opts: Opts{
				Name:           "foo",
				Help:           "Some help.",
				Labels:         Labels{"foo": "bar"},
				VariableLabels: []string{"baz:baz"},
			},
			scalarOK: false,
			vecOK:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if tt.scalarOK {
				assertScalarOptsOK(t, tt.opts)
			} else {
				assertScalarOptsFail(t, tt.opts)
			}
			if tt.vecOK {
				assertVectorOptsOK(t, tt.opts)
			} else {
				assertVectorOptsFail(t, tt.opts)
			}
		})
	}
}

func TestHistogramOptsValidation(t *testing.T) {
	tests := []struct {
		desc     string
		opts     HistogramOpts
		scalarOK bool
		vecOK    bool
	}{
		{
			desc: "valid names",
			opts: HistogramOpts{
				Opts: Opts{
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
			desc: "valid names & constant labels",
			opts: HistogramOpts{
				Opts: Opts{
					Name:   "foo",
					Help:   "Some help.",
					Labels: Labels{"foo": "bar"},
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: true,
			vecOK:    false,
		},
		{
			desc: "name with Tally-forbidden characters",
			opts: HistogramOpts{
				Opts: Opts{
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
			opts: HistogramOpts{
				Opts: Opts{
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
			opts: HistogramOpts{
				Opts: Opts{
					Name: "foo",
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: false,
			vecOK:    false,
		},
		{
			desc: "valid names but invalid label key",
			opts: HistogramOpts{
				Opts: Opts{
					Name:   "foo",
					Help:   "Some help.",
					Labels: Labels{"foo:foo": "bar"},
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: true,
			vecOK:    false,
		},
		{
			desc: "valid names but invalid label value",
			opts: HistogramOpts{
				Opts: Opts{
					Name:   "foo",
					Help:   "Some help.",
					Labels: Labels{"foo": "bar:bar"},
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: true,
			vecOK:    false,
		},
		{
			desc: "valid names & variable labels",
			opts: HistogramOpts{
				Opts: Opts{
					Name:           "foo",
					Help:           "Some help.",
					VariableLabels: []string{"baz"},
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: false,
			vecOK:    true,
		},
		{
			desc: "valid names, constant labels, & variable labels",
			opts: HistogramOpts{
				Opts: Opts{
					Name:           "foo",
					Help:           "Some help.",
					Labels:         Labels{"foo": "bar"},
					VariableLabels: []string{"baz"},
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: false,
			vecOK:    true,
		},
		{
			desc: "valid names & constant labels, but invalid variable labels",
			opts: HistogramOpts{
				Opts: Opts{
					Name:           "foo",
					Help:           "Some help.",
					Labels:         Labels{"foo": "bar"},
					VariableLabels: []string{"baz:baz"},
				},
				Unit:    time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: false,
			vecOK:    true,
		},
		{
			desc: "valid labels, no unit",
			opts: HistogramOpts{
				Opts: Opts{
					Name:           "foo",
					Help:           "Some help.",
					Labels:         Labels{"foo": "bar"},
					VariableLabels: []string{"baz"},
				},
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: false,
			vecOK:    false,
		},
		{
			desc: "valid labels, negative unit",
			opts: HistogramOpts{
				Opts: Opts{
					Name:           "foo",
					Help:           "Some help.",
					Labels:         Labels{"foo": "bar"},
					VariableLabels: []string{"baz"},
				},
				Unit:    -1 * time.Millisecond,
				Buckets: []int64{1000, 1000 * 60},
			},
			scalarOK: false,
			vecOK:    false,
		},
		{
			desc: "valid labels, no buckets",
			opts: HistogramOpts{
				Opts: Opts{
					Name:           "foo",
					Help:           "Some help.",
					Labels:         Labels{"foo": "bar"},
					VariableLabels: []string{"baz"},
				},
				Unit: time.Millisecond,
			},
			scalarOK: false,
			vecOK:    false,
		},
		{
			desc: "valid labels, buckets out of order",
			opts: HistogramOpts{
				Opts: Opts{
					Name:           "foo",
					Help:           "Some help.",
					Labels:         Labels{"foo": "bar"},
					VariableLabels: []string{"baz"},
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
				assertScalarHistogramOptsOK(t, tt.opts)
			} else {
				assertSimpleHistogramOptsFail(t, tt.opts)
			}
			if tt.vecOK {
				assertVectorHistogramOptsOK(t, tt.opts)
			} else {
				assertVectorHistogramOptsFail(t, tt.opts)
			}
		})
	}
}

func justRegistry(r *Registry, _ *Controller) *Registry {
	return r
}

func assertScalarOptsOK(t testing.TB, opts Opts) {
	_, err := justRegistry(New()).NewCounter(opts)
	assert.NoError(t, err, "Expected success from NewCounter.")

	_, err = justRegistry(New()).NewGauge(opts)
	assert.NoError(t, err, "Expected success from NewGauge.")
}

func assertScalarOptsFail(t testing.TB, opts Opts) {
	_, err := justRegistry(New()).NewCounter(opts)
	assert.Error(t, err, "Expected an error from NewCounter.")

	_, err = justRegistry(New()).NewGauge(opts)
	assert.Error(t, err, "Expected an error from NewGauge.")
}

func assertVectorOptsOK(t testing.TB, opts Opts) {
	_, err := justRegistry(New()).NewCounterVector(opts)
	assert.NoError(t, err, "Expected success from NewCounterVector.")

	_, err = justRegistry(New()).NewGaugeVector(opts)
	assert.NoError(t, err, "Expected success from NewGaugeVector.")
}

func assertVectorOptsFail(t testing.TB, opts Opts) {
	_, err := justRegistry(New()).NewCounterVector(opts)
	assert.Error(t, err, "Expected an error from NewCounterVector.")

	_, err = justRegistry(New()).NewGaugeVector(opts)
	assert.Error(t, err, "Expected an error from NewGaugeVector.")
}

func assertScalarHistogramOptsOK(t testing.TB, opts HistogramOpts) {
	_, err := justRegistry(New()).NewHistogram(opts)
	assert.NoError(t, err, "Expected success from NewLatencies.")
}

func assertSimpleHistogramOptsFail(t testing.TB, opts HistogramOpts) {
	_, err := justRegistry(New()).NewHistogram(opts)
	assert.Error(t, err, "Expected an error from NewLatencies.")
}

func assertVectorHistogramOptsOK(t testing.TB, opts HistogramOpts) {
	_, err := justRegistry(New()).NewHistogramVector(opts)
	assert.NoError(t, err, "Expected success from NewLatenciesVector.")
}

func assertVectorHistogramOptsFail(t testing.TB, opts HistogramOpts) {
	_, err := justRegistry(New()).NewHistogramVector(opts)
	assert.Error(t, err, "Expected an error from NewLatenciesVector.")
}
