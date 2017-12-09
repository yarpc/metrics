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
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"

	. "go.uber.org/net/metrics"
	"go.uber.org/net/metrics/tallypush"
)

func initializeMetrics(t testing.TB, disablePush bool) *Root {
	root := New()
	scope := root.Scope().Tagged(Tags{"service": "users"})

	counter, err := scope.Counter(Spec{
		Name:        "test_counter",
		Help:        "counter help",
		ConstTags:   Tags{"foo": "counter"},
		DisablePush: disablePush,
	})
	require.NoError(t, err, "Failed to create counter.")
	counter.Inc()

	counterVec, err := scope.CounterVector(Spec{
		Name:        "test_counter_vector",
		Help:        "counter vector help",
		ConstTags:   Tags{"foo": "counter_vector"},
		VarTags:     []string{"quux", "baz"},
		DisablePush: disablePush,
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

	gauge, err := scope.Gauge(Spec{
		Name:        "test_gauge",
		Help:        "gauge help",
		ConstTags:   Tags{"foo": "gauge"},
		DisablePush: disablePush,
	})
	require.NoError(t, err, "Failed to create gauge.")
	gauge.Store(42)

	gaugeVec, err := scope.GaugeVector(Spec{
		Name:        "test_gauge_vector",
		Help:        "gauge vector help",
		ConstTags:   Tags{"foo": "gauge_vector"},
		VarTags:     []string{"quux", "baz"},
		DisablePush: disablePush,
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

	hist, err := scope.Histogram(HistogramSpec{
		Spec: Spec{
			Name:        "test_histogram",
			Help:        "histogram help",
			ConstTags:   Tags{"foo": "histogram"},
			DisablePush: disablePush,
		},
		Unit:    time.Millisecond,
		Buckets: []int64{1000, 1000 * 60},
	})
	require.NoError(t, err, "Failed to create histogram.")
	hist.Observe(time.Millisecond)

	histVec, err := scope.HistogramVector(HistogramSpec{
		Spec: Spec{
			Name:        "test_histogram_vector",
			Help:        "histogram vector help",
			ConstTags:   Tags{"foo": "histogram_vector"},
			VarTags:     []string{"quux", "baz"},
			DisablePush: disablePush,
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

	return root
}

func snapshot(t testing.TB, root *Root) tally.Snapshot {
	tallyScope := tally.NewTestScope("" /* prefix */, nil /* tags */)
	stop, err := root.Push(tallypush.New(tallyScope), 10*time.Millisecond)
	require.NoError(t, err, "Couldn't start Tally push.")

	_, err = root.Push(tallypush.New(tallyScope), 10*time.Millisecond)
	require.Error(t, err, "Shoudn't be able to run multiple push goroutines concurrently.")

	time.Sleep(100 * time.Millisecond)
	stop()

	return tallyScope.Snapshot()
}

func TestTallyEndToEnd(t *testing.T) {
	// Since the metric name and tags are encoded into the Tally snapshot map
	// keys, we're only going to explicitly assert the values.

	t.Run("export disabled", func(t *testing.T) {
		root := initializeMetrics(t, true)
		snap := snapshot(t, root)
		assert.Zero(
			t,
			len(snap.Timers())+len(snap.Counters())+len(snap.Gauges())+len(snap.Histograms()),
			"Shouldn't export any metrics.",
		)
	})

	t.Run("export enabled", func(t *testing.T) {
		root := initializeMetrics(t, false)
		snap := snapshot(t, root)
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

// scrape collects and returns the plain-text content of a GET on the supplied
// handler, along with the response code.
func scrape(t testing.TB, handler http.Handler) (int, string) {
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err, "Unexpected error scraping Prometheus endpoint.")
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err, "Unexpected error reading response body.")
	return resp.StatusCode, strings.TrimSpace(string(body))
}

// assertPrometheus asserts that the root's scrape endpoint successfully
// serves the supplied plain-text Prometheus metrics.
func assertPrometheus(t testing.TB, root *Root, expected string) {
	code, actual := scrape(t, root)
	assert.Equal(t, http.StatusOK, code, "Unexpected HTTP response code from Prometheus scrape.")
	assert.Equal(
		t,
		strings.Split(expected, "\n"),
		strings.Split(actual, "\n"),
		"Unexpected Prometheus text.",
	)
}

func TestPrometheusEndToEnd(t *testing.T) {
	root := initializeMetrics(t, false)
	// This fixture was generated by the vanilla Prometheus client. Keeping this
	// test passing verifies that data exposed by net/metrics is
	// indistinguishable from data exposed by the official Prometheus clients.
	bs, err := ioutil.ReadFile("testdata/proto_integration_test.txt")
	require.NoError(t, err, "Failed to open test fixture.")
	assertPrometheus(t, root, string(bs))
}
