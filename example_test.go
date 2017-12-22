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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"time"

	"github.com/uber-go/tally"
	"go.uber.org/net/metrics"
	"go.uber.org/net/metrics/tallypush"
)

func Example() {
	// First, construct a metrics root. Generally, there's only one root in each
	// process.
	root := metrics.New()
	// From the root, access the top-level scope and add some tags to create a
	// sub-scope. You'll typically pass scopes around your application, since
	// they let you create individual metrics.
	scope := root.Scope().Tagged(metrics.Tags{
		"host":   "db01",
		"region": "us-west",
	})

	// Create a simple counter. Note that the name is fairly long; if this code
	// were part of a reusable library called "foo", "foo_selects_completed"
	// would be a much better name.
	total, err := scope.Counter(metrics.Spec{
		Name: "selects_completed",
		Help: "Total number of completed SELECT queries.",
	})
	if err != nil {
		panic(err)
	}

	// See the package-level documentation for a general discussion of vectors.
	// In this case, we're going to track the number of in-progress SELECT
	// queries by table and user. Since we won't know the table and user names
	// until we actually receive each query, we model this as a vector with two
	// variable tags.
	progress, err := scope.GaugeVector(metrics.Spec{
		Name:    "selects_in_progress",
		Help:    "Number of in-progress SELECT queries.",
		VarTags: []string{"table", "user"},
	})
	if err != nil {
		panic(err)
	}
	// MustGet retrieves the gauge with the specified variable tags, creating
	// one if necessary. We must supply both the variable tag names and values,
	// and they must be in the correct order. MustGet panics only if the tags
	// are malformed. If you'd rather check errors explicitly, there's also a
	// Get method.
	trips := progress.MustGet(
		"table" /* tag name */, "trips", /* tag value */
		"user" /* tag name */, "jane", /* tag value */
	)
	drivers := progress.MustGet(
		"table", "drivers",
		"user", "chen",
	)

	fmt.Println("Trips:", trips.Inc())
	total.Inc()
	fmt.Println("Drivers:", drivers.Add(2))
	total.Add(2)
	fmt.Println("Drivers:", drivers.Dec())
	fmt.Println("Trips:", trips.Dec())
	fmt.Println("Total:", total.Load())

	// Output:
	// Trips: 1
	// Drivers: 2
	// Drivers: 1
	// Trips: 0
	// Total: 3
}

func ExampleCounter() {
	c, err := metrics.New().Scope().Counter(metrics.Spec{
		Name:      "selects_completed",                         // required
		Help:      "Total number of completed SELECT queries.", // required
		ConstTags: metrics.Tags{"host": "db01"},                // optional
	})
	if err != nil {
		panic(err)
	}
	c.Add(2)
}

func ExampleCounterVector() {
	vec, err := metrics.New().Scope().CounterVector(metrics.Spec{
		Name:      "selects_completed_by_table",                   // required
		Help:      "Number of completed SELECT queries by table.", // required
		ConstTags: metrics.Tags{"host": "db01"},                   // optional
		VarTags:   []string{"table"},                              // required
	})
	if err != nil {
		panic(err)
	}
	vec.MustGet("table" /* tag name */, "trips" /* tag value */).Inc()
}

func ExampleGauge() {
	g, err := metrics.New().Scope().Gauge(metrics.Spec{
		Name:      "selects_in_progress",                       // required
		Help:      "Total number of in-flight SELECT queries.", // required
		ConstTags: metrics.Tags{"host": "db01"},                // optional
	})
	if err != nil {
		panic(err)
	}
	g.Store(11)
}

func ExampleGaugeVector() {
	vec, err := metrics.New().Scope().GaugeVector(metrics.Spec{
		Name:      "selects_in_progress_by_table",                 // required
		Help:      "Number of in-flight SELECT queries by table.", // required
		ConstTags: metrics.Tags{"host": "db01"},                   // optional
		VarTags:   []string{"table"},                              // optional
	})
	if err != nil {
		panic(err)
	}
	vec.MustGet("table" /* tag name */, "trips" /* tag value */).Store(11)
}

func ExampleHistogram() {
	h, err := metrics.New().Scope().Histogram(metrics.HistogramSpec{
		Spec: metrics.Spec{
			Name:      "selects_latency_ms",         // required, should indicate unit
			Help:      "SELECT query latency.",      // required
			ConstTags: metrics.Tags{"host": "db01"}, // optional
		},
		Unit:    time.Millisecond,                      // required
		Buckets: []int64{5, 10, 25, 50, 100, 200, 500}, // required
	})
	if err != nil {
		panic(err)
	}
	h.Observe(37 * time.Millisecond) // increments bucket with upper bound 50
	h.IncBucket(37)                  // also increments bucket with upper bound 50
}

func ExampleHistogramVector() {
	vec, err := metrics.New().Scope().HistogramVector(metrics.HistogramSpec{
		Spec: metrics.Spec{
			Name:      "selects_latency_by_table_ms",    // required, should indicate unit
			Help:      "SELECT query latency by table.", // required
			ConstTags: metrics.Tags{"host": "db01"},     // optional
			VarTags:   []string{"table"},
		},
		Unit:    time.Millisecond,                      // required
		Buckets: []int64{5, 10, 25, 50, 100, 200, 500}, // required
	})
	if err != nil {
		panic(err)
	}
	vec.MustGet("table" /* tag name */, "trips" /* tag value */).Observe(37 * time.Millisecond)
}

func ExampleRoot_ServeHTTP() {
	// First, construct a root and add some metrics.
	root := metrics.New()
	c, err := root.Scope().Counter(metrics.Spec{
		Name:      "example",
		Help:      "Counter demonstrating HTTP exposition.",
		ConstTags: metrics.Tags{"host": "example01"},
	})
	if err != nil {
		panic(err)
	}
	c.Inc()

	// Expose the root on your HTTP server of choice.
	mux := http.NewServeMux()
	mux.Handle("/debug/net/metrics", root)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Your metrics are now exposed via a Prometheus-compatible handler. This
	// example shows text output, but clients can also request the protocol
	// buffer binary format.
	res, err := http.Get(fmt.Sprintf("%v/debug/net/metrics", srv.URL))
	if err != nil {
		panic(err)
	}
	text, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(text))

	// Output:
	// # HELP example Counter demonstrating HTTP exposition.
	// # TYPE example counter
	// example{host="example01"} 1
}

func ExampleRoot_Push() {
	// First, we need something to push to. In this example, we'll use Tally's
	// testing scope.
	ts := tally.NewTestScope("" /* prefix */, nil /* tags */)
	root := metrics.New()

	// Push updates to our test scope twice per second.
	stop, err := root.Push(tallypush.New(ts), 500*time.Millisecond)
	if err != nil {
		panic(err)
	}
	defer stop()

	c, err := root.Scope().Counter(metrics.Spec{
		Name: "example",
		Help: "Counter demonstrating push integration.",
	})
	if err != nil {
		panic(err)
	}
	c.Inc()

	// Sleep to make sure that we run at least one push, then print the counter
	// value as seen by Tally.
	time.Sleep(2 * time.Second)
	fmt.Println(ts.Snapshot().Counters()["example+"].Value())

	// Output:
	// 1
}

func ExampleRoot_Snapshot() {
	// Snapshots are the simplest way to unit test your metrics. A future
	// release will add a more full-featured metricstest package.
	root := metrics.New()
	c, err := root.Scope().Counter(metrics.Spec{
		Name:      "example",
		Help:      "Counter demonstrating snapshots.",
		ConstTags: metrics.Tags{"foo": "bar"},
	})
	if err != nil {
		panic(err)
	}
	c.Inc()

	// It's safe to snapshot your metrics in production, but keep in mind that
	// taking a snapshot is relatively slow and expensive.
	actual := root.Snapshot().Counters[0]
	expected := metrics.Snapshot{
		Name:  "example",
		Value: 1,
		Tags:  metrics.Tags{"foo": "bar"},
	}
	if !reflect.DeepEqual(expected, actual) {
		panic(fmt.Sprintf("expected %v, got %v", expected, actual))
	}
}
