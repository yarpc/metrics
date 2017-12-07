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

// A Counter is a monotonically increasing value, like a car's odometer. All
// its exported methods are safe to use concurrently.
//
// Nil *Counters are safe no-op implementations.
type Counter struct {
	meta metadata
}

func newCounter(m metadata) *Counter {
	return &Counter{m}
}

// Add increases the value of the counter and returns the new value. Since
// counters must be monotonically increasing, passing a negative number just
// returns the current value (without modifying it).
func (c *Counter) Add(n int64) int64 {
	if c == nil {
		return 0
	}
	return 0
}

// Inc increments the counter's value by one and returns the new value.
func (c *Counter) Inc() int64 {
	if c == nil {
		return 0
	}
	return 0
}

// Load returns the counter's current value.
func (c *Counter) Load() int64 {
	if c == nil {
		return 0
	}
	return 0
}

func (c *Counter) describe() metadata {
	return c.meta
}

// A CounterVector is a collection of Counters that share a name and some
// constant labels, but also have a consistent set of variable labels.
// All exported methods are safe to use concurrently.
//
// A nil *CounterVector is safe to use, and always returns no-op counters.
//
// For a general description of vector types, see the package-level
// documentation.
type CounterVector struct {
	meta metadata
}

func newCounterVector(m metadata) *CounterVector {
	return &CounterVector{m}
}

// Get retrieves the counter with the supplied variable label names and values
// from the vector, creating one if necessary. The variable labels must be
// supplied in the same order used when creating the vector.
//
// Get returns an error if the number or order of labels is incorrect.
func (cv *CounterVector) Get(variableLabels ...string) (*Counter, error) {
	if cv == nil {
		return nil, nil
	}
	return nil, nil
}

// MustGet behaves exactly like Get, but panics on errors. If code using this
// method is covered by unit tests, this is safe.
func (cv *CounterVector) MustGet(variableLabels ...string) *Counter {
	if cv == nil {
		return nil
	}
	c, err := cv.Get(variableLabels...)
	if err != nil {
		panic(fmt.Sprintf("failed to get counter: %v", err))
	}
	return c
}

func (cv *CounterVector) describe() metadata {
	return cv.meta
}
