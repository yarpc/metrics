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
	"fmt"
)

// A Gauge is a point-in-time measurement, like a car's speedometer. All its
// exported methods are safe to use concurrently.
//
// Nil *Gauges are safe no-op implementations.
type Gauge struct {
}

// Add increases the value of the gauge and returns the new value. Adding
// negative values is allowed, but using Sub may be simpler.
func (g *Gauge) Add(n int64) int64 {
	if g == nil {
		return 0
	}
	return 0
}

// Sub decreases the value of the gauge and returns the new value. Subtracting
// negative values is allowed, but using Add may be simpler.
func (g *Gauge) Sub(n int64) int64 {
	if g == nil {
		return 0
	}
	return 0
}

// Inc increments the gauge's current value by one and returns the new value.
func (g *Gauge) Inc() int64 {
	return g.Add(1)
}

// Dec decrements the gauge's current value by one and returns the new value.
func (g *Gauge) Dec() int64 {
	return g.Sub(1)
}

// Swap replaces the gauge's current value and returns the previous value.
func (g *Gauge) Swap(n int64) int64 {
	if g == nil {
		return 0
	}
	return 0
}

// CAS is a compare-and-swap. It compares the current value to the old value
// supplied, and if they match it stores the new value. The return value
// indicates whether the swap succeeded.
func (g *Gauge) CAS(old, new int64) bool {
	if g == nil {
		return false
	}
	return false
}

// Store changes the gauge's value.
func (g *Gauge) Store(n int64) {
}

// Load returns the gauge's current value.
func (g *Gauge) Load() int64 {
	if g == nil {
		return 0
	}
	return 0
}

// A GaugeVector is a collection of Gauges that share a name and some constant
// labels, but also have a consistent set of variable labels. All exported
// methods are safe to use concurrently.
//
// A nil *GaugeVector is safe to use, and always returns no-op gauges.
//
// For a general description of vector types, see the package-level
// documentation.
type GaugeVector struct {
}

// Get retrieves the gauge with the supplied variable label names and values
// from the vector, creating one if necessary. The variable labels must be
// supplied in the same order used when creating the vector.
//
// Get returns an error if the number or order of labels is incorrect.
func (gv *GaugeVector) Get(variableLabels ...string) (*Gauge, error) {
	if gv == nil {
		return nil, nil
	}
	return nil, nil
}

// MustGet behaves exactly like Get, but panics on errors. If code using this
// method is covered by unit tests, this is safe.
func (gv *GaugeVector) MustGet(variableLabels ...string) *Gauge {
	if gv == nil {
		return nil
	}
	g, err := gv.Get(variableLabels...)
	if err != nil {
		panic(fmt.Sprintf("failed to get gauge: %v", err))
	}
	return g
}
