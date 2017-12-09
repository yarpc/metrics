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
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/net/metrics/push"

	"go.uber.org/atomic"
)

// An Option configures a Root. Currently, there are no exported Options.
type Option interface {
	unimplemented()
}

// A Root is a collection of tagged metrics that can be exposed via in-memory
// snapshots, push-based telemetry systems, or a Prometheus-compatible HTTP
// handler.
type Root struct {
	*core

	scope   *Scope
	pushing atomic.Bool // can only push to one target
	handler http.Handler
}

// New constructs a Root.
func New(opts ...Option) *Root {
	core := newCore()
	return &Root{
		core:  core,
		scope: newScope(core, Tags{}),
		handler: promhttp.HandlerFor(core.gatherer, promhttp.HandlerOpts{
			ErrorHandling: promhttp.HTTPErrorOnError, // 500 on errors
		}),
	}
}

// Scope exposes the Root's top-level metrics collection. Tagged sub-scopes
// and individual counters, gauges, histograms, and vectors can be created
// from this top-level Scope.
func (r *Root) Scope() *Scope {
	return r.scope
}

// ServeHTTP implements http.Handler.
func (r *Root) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}

// Snapshot returns a point-in-time view of all the metrics contained in the
// root (and all its scopes). It's safe to use concurrently, but is relatively
// expensive and designed for use in unit tests.
func (r *Root) Snapshot() *RootSnapshot {
	return r.snapshot()
}

// Push starts a goroutine that periodically exports all registered metrics to
// the supplied target. Controllers may only push to a single target at a
// time; to push to multiple backends simultaneously, implement a teeing
// push.Target.
//
// The returned function cleanly shuts down the background goroutine.
func (r *Root) Push(target push.Target, tick time.Duration) (context.CancelFunc, error) {
	if r.pushing.Swap(true) {
		return nil, errors.New("already pushing")
	}
	pusher := newPusher(r.core, target, tick)
	go pusher.Start()
	// We don't want to set c.pushing to false when we stop the push loop,
	// because that would let users start another pusher. Since pushers are
	// usually stateful, this would immediately re-push all the counter
	// increments since process startup.
	return pusher.Stop, nil
}
