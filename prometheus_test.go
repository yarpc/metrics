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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.uber.org/net/metrics"
)

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

func TestZeroLengthVectors(t *testing.T) {
	root := New()
	root.Scope().CounterVector(Spec{
		Name:    "counter_vector",
		Help:    "some help",
		VarTags: []string{"vary"},
	})
	root.Scope().GaugeVector(Spec{
		Name:    "gauge_vector",
		Help:    "some help",
		VarTags: []string{"vary"},
	})
	root.Scope().HistogramVector(HistogramSpec{
		Spec: Spec{
			Name:    "gauge_vector",
			Help:    "some help",
			VarTags: []string{"vary"},
		},
		Unit:    time.Millisecond,
		Buckets: []int64{1, 2, 3},
	})
	code, body := scrape(t, root)
	assert.Equal(t, http.StatusOK, code, "Expected GET to succeed.")
	assert.Zero(t, body, "Expected empty body.")
}
