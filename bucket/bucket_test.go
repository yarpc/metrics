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

package bucket

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type int64s []int64

func (is int64s) Len() int           { return len(is) }
func (is int64s) Less(i, j int) bool { return is[i] < is[j] }
func (is int64s) Swap(i, j int)      { is[i], is[j] = is[j], is[i] }

func assertStrictlySorted(t testing.TB, is []int64) {
	// We want the same result as the production code's isAscending without
	// copying logic.
	if !sort.IsSorted(int64s(is)) {
		t.Logf("not sorted: %v", is)
		t.Fail()
	}
	if hasDuplicates(is) {
		t.Logf("has duplicate buckets: %v", is)
		t.Fail()
	}
}

func hasDuplicates(is []int64) bool {
	set := make(map[int64]struct{}, len(is))
	for _, i := range is {
		if _, ok := set[i]; ok {
			return true
		}
		set[i] = struct{}{}
	}
	return false
}

func TestNewRPC(t *testing.T) {
	bs := NewRPCLatency()
	require.True(t, len(bs) > 0, "Expected at least one bucket.")
	assertStrictlySorted(t, bs)
}

func TestNewExponential(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bs := NewExponential(1, 2, 3)
		assert.Equal(t, []int64{1, 2, 4}, bs)
	})
	t.Run("too few buckets", func(t *testing.T) {
		bs := NewExponential(1, 2, 0)
		assert.Equal(t, 0, len(bs))
	})
	t.Run("negative start", func(t *testing.T) {
		bs := NewExponential(-1, 2, 3)
		assert.Equal(t, 0, len(bs))
	})
	t.Run("zero start", func(t *testing.T) {
		bs := NewExponential(0, 2, 3)
		assert.Equal(t, 0, len(bs))
	})
	t.Run("negative factor", func(t *testing.T) {
		bs := NewExponential(1, -2, 3)
		assert.Equal(t, 0, len(bs))
	})
	t.Run("zero factor", func(t *testing.T) {
		bs := NewExponential(1, 0, 3)
		assert.Equal(t, 0, len(bs))
	})
}

func TestNewLinear(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bs := NewLinear(-2, 2, 4)
		assert.Equal(t, []int64{-2, 0, 2, 4}, bs)
	})
	t.Run("too few buckets", func(t *testing.T) {
		bs := NewLinear(2, 2, 0)
		assert.Equal(t, 0, len(bs))
	})
	t.Run("negative width", func(t *testing.T) {
		bs := NewLinear(-2, -1, 4)
		assert.Equal(t, 0, len(bs))
	})
	t.Run("zero width", func(t *testing.T) {
		bs := NewLinear(-2, 0, 4)
		assert.Equal(t, 0, len(bs))
	})
}

func TestFlatten(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bs, err := Flatten([]int64{1, 2, 3}, []int64{5, 10, 15})
		require.NoError(t, err, "Failed to flatten buckets.")
		assert.Equal(t, []int64{1, 2, 3, 5, 10, 15}, bs)
	})
	t.Run("subslice not sorted", func(t *testing.T) {
		_, err := Flatten([]int64{1, 3, 2}, []int64{5, 10, 15})
		require.Error(t, err)
	})
	t.Run("subslice not uniqued", func(t *testing.T) {
		_, err := Flatten([]int64{1, 2, 2}, []int64{5, 10, 15})
		require.Error(t, err)
	})
	t.Run("slices overlap", func(t *testing.T) {
		_, err := Flatten([]int64{1, 2, 7}, []int64{5, 10, 15})
		require.Error(t, err)
	})
	t.Run("slices share bounds", func(t *testing.T) {
		_, err := Flatten([]int64{1, 2, 5}, []int64{5, 10, 15})
		require.Error(t, err)
	})
}
