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

// Controller has the ability to expose the metrics that are registered against
// the registry this controller was created with.
type Controller struct {
	*coreRegistry

	pushing atomic.Bool // can only push to one target
	handler http.Handler
}

func newController(core *coreRegistry) *Controller {
	return &Controller{
		coreRegistry: core,
		handler: promhttp.HandlerFor(core.gatherer, promhttp.HandlerOpts{
			ErrorHandling: promhttp.HTTPErrorOnError, // 500 on errors
		}),
	}
}

// ServeHTTP implements http.Handler.
func (c *Controller) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c.handler.ServeHTTP(w, req)
}

// Snapshot returns a point-in-time view of all the metrics contained in the
// controller's registry. It's safe to use concurrently, but is relatively
// expensive and designed for use in unit tests.
func (c *Controller) Snapshot() *RegistrySnapshot {
	return c.snapshot()
}

// Push starts a goroutine that periodically exports all registered metrics to
// the supplied target. Controllers may only push to a single target at a
// time; to push to multiple backends simultaneously, implement a teeing
// push.Target.
//
// The returned function cleanly shuts down the background goroutine.
func (c *Controller) Push(target push.Target, tick time.Duration) (context.CancelFunc, error) {
	if c.pushing.Swap(true) {
		return nil, errors.New("already pushing")
	}
	pusher := newPusher(c.coreRegistry, target, tick)
	go pusher.Start()
	// We don't want to set c.pushing to false when we stop the push loop,
	// because that would let users start another pusher. Since pushers are
	// usually stateful, this would immediately re-push all the counter
	// increments since process startup.
	return pusher.Stop, nil
}
