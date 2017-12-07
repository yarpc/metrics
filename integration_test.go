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

package metrics_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"

	. "go.uber.org/net/metrics"
	"go.uber.org/net/metrics/tallypush"
)

func initializeMetrics(t testing.TB, disablePush bool) *Controller {
	root, controller := New()
	reg := root.Labeled(Labels{"service": "users"})

	counter, err := reg.NewCounter(Opts{
		Name:        "test_counter",
		Help:        "counter help",
		Labels:      Labels{"foo": "counter"},
		DisablePush: disablePush,
	})
	require.NoError(t, err, "Failed to create counter.")
	counter.Inc()

	counterVec, err := reg.NewCounterVector(Opts{
		Name:           "test_counter_vector",
		Help:           "counter vector help",
		Labels:         Labels{"foo": "counter_vector"},
		VariableLabels: []string{"quux", "baz"},
		DisablePush:    disablePush,
	})
	require.NoError(t, err, "Failed to create counter vector.")
	counterVec.MustGet(
		"quux", "quuxval",
		"baz", "bazval",
	).Inc()
	counterVec.MustGet(
		"quux", "quuxval2",
		"baz", "bazval2",
	).Inc()

	gauge, err := reg.NewGauge(Opts{
		Name:        "test_gauge",
		Help:        "gauge help",
		Labels:      Labels{"foo": "gauge"},
		DisablePush: disablePush,
	})
	require.NoError(t, err, "Failed to create gauge.")
	gauge.Store(42)

	gaugeVec, err := reg.NewGaugeVector(Opts{
		Name:           "test_gauge_vector",
		Help:           "gauge vector help",
		Labels:         Labels{"foo": "gauge_vector"},
		VariableLabels: []string{"quux", "baz"},
		DisablePush:    disablePush,
	})
	require.NoError(t, err, "Failed to create gauge vector.")
	gaugeVec.MustGet(
		"quux", "quuxval",
		"baz", "bazval",
	).Store(10)
	gaugeVec.MustGet(
		"quux", "quuxval2",
		"baz", "bazval2",
	).Store(20)

	hist, err := reg.NewHistogram(HistogramOpts{
		Opts: Opts{
			Name:        "test_histogram",
			Help:        "histogram help",
			Labels:      Labels{"foo": "histogram"},
			DisablePush: disablePush,
		},
		Unit:    time.Millisecond,
		Buckets: []int64{1000, 1000 * 60},
	})
	require.NoError(t, err, "Failed to create histogram.")
	hist.Observe(time.Millisecond)

	histVec, err := reg.NewHistogramVector(HistogramOpts{
		Opts: Opts{
			Name:           "test_histogram_vector",
			Help:           "histogram vector help",
			Labels:         Labels{"foo": "histogram_vector"},
			VariableLabels: []string{"quux", "baz"},
			DisablePush:    disablePush,
		},
		Unit:    time.Millisecond,
		Buckets: []int64{1000, 1000 * 60},
	})
	require.NoError(t, err, "Failed to create histogram vector.")
	histVec.MustGet(
		"quux", "quuxval",
		"baz", "bazval",
	).Observe(time.Millisecond)
	histVec.MustGet(
		"quux", "quuxval2",
		"baz", "bazval2",
	).Observe(time.Millisecond)

	return controller
}

func snapshot(t testing.TB, controller *Controller) tally.Snapshot {
	scope := tally.NewTestScope("" /* prefix */, nil /* tags */)
	stop, err := controller.Push(tallypush.New(scope), 10*time.Millisecond)
	require.NoError(t, err, "Couldn't start Tally push.")

	_, err = controller.Push(tallypush.New(scope), 10*time.Millisecond)
	require.Error(t, err, "Shoudn't be able to run multiple push goroutines concurrently.")

	time.Sleep(100 * time.Millisecond)
	stop()

	return scope.Snapshot()
}

func TestTallyEndToEnd(t *testing.T) {
	// Since the metric name and tags are encoded into the Tally snapshot map
	// keys, we're only going to explicitly assert the values.

	t.Run("export disabled", func(t *testing.T) {
		controller := initializeMetrics(t, true)
		snap := snapshot(t, controller)
		assert.Zero(
			t,
			len(snap.Timers())+len(snap.Counters())+len(snap.Gauges())+len(snap.Histograms()),
			"Shouldn't export any metrics.",
		)
	})

	t.Run("export enabled", func(t *testing.T) {
		controller := initializeMetrics(t, false)
		snap := snapshot(t, controller)
		assert.Zero(t, len(snap.Timers()), "Shouldn't export any timers.")

		counters := snap.Counters()
		assert.Equal(t, 3, len(counters), "Wrong number of counters.")
		assert.Equal(t,
			int64(1),
			counters["test_counter+foo=counter,service=users"].Value(),
			"Wrong value for scalar counter.",
		)
		assert.Equal(t,
			int64(1),
			counters["test_counter_vector+baz=bazval,foo=counter_vector,quux=quuxval,service=users"].Value(),
			"Wrong value for first vectorized counter.",
		)
		assert.Equal(t,
			int64(1),
			counters["test_counter_vector+baz=bazval2,foo=counter_vector,quux=quuxval2,service=users"].Value(),
			"Wrong value for second vectorized counter.",
		)

		gauges := snap.Gauges()
		assert.Equal(t, 3, len(gauges), "Wrong number of gauges.")
		assert.Equal(t,
			float64(42),
			gauges["test_gauge+foo=gauge,service=users"].Value(),
			"Wrong value for scalar gauge.",
		)
		assert.Equal(t,
			float64(10),
			gauges["test_gauge_vector+baz=bazval,foo=gauge_vector,quux=quuxval,service=users"].Value(),
			"Wrong value for first vectorized gauge.",
		)
		assert.Equal(t,
			float64(20),
			gauges["test_gauge_vector+baz=bazval2,foo=gauge_vector,quux=quuxval2,service=users"].Value(),
			"Wrong value for second vectorized gauge.",
		)

		histograms := snap.Histograms()
		assert.Equal(t, 3, len(histograms), "Wrong number of histograms.")
		assert.Equal(t,
			map[float64]int64{1000: 1, 1000 * 60: 0, math.MaxFloat64: 0},
			histograms["test_histogram+foo=histogram,service=users"].Values(),
			"Wrong value for scalar histogram.",
		)
		assert.Equal(t,
			map[float64]int64{1000: 1, 1000 * 60: 0, math.MaxFloat64: 0},
			histograms["test_histogram_vector+baz=bazval,foo=histogram_vector,quux=quuxval,service=users"].Values(),
			"Wrong value for first vectorized histogram.",
		)
		assert.Equal(t,
			map[float64]int64{1000: 1, 1000 * 60: 0, math.MaxFloat64: 0},
			histograms["test_histogram_vector+baz=bazval2,foo=histogram_vector,quux=quuxval2,service=users"].Values(),
			"Wrong value for second vectorized histogram.",
		)
	})
}
