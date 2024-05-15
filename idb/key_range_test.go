//go:build js && wasm
// +build js,wasm

package idb

import (
	"fmt"
	"syscall/js"
	"testing"

	"github.com/aperturerobotics/go-indexeddb/idb/internal/assert"
	"github.com/hack-pad/safejs"
)

func TestNewKeyRangeBound(t *testing.T) {
	t.Parallel()
	keyRangeClosedOpen, err := NewKeyRangeBound(safejs.Safe(js.ValueOf(0)), safejs.Safe(js.ValueOf(100)), false, true)
	assert.NoError(t, err)
	for _, tc := range []struct {
		name           string // auto-filled, satisfies paralleltest linter
		input          int
		expectIncludes bool
	}{
		{input: -1, expectIncludes: false},
		{input: 0, expectIncludes: true},
		{input: 50, expectIncludes: true},
		{input: 100, expectIncludes: false},
	} {
		tc.name = fmt.Sprint("closed open ", tc.input)
		tc := tc // keep loop-local copy of test case for parallel runs
		t.Run(fmt.Sprint(tc.input), func(t *testing.T) {
			t.Parallel()
			includes, err := keyRangeClosedOpen.Includes(safejs.Safe(js.ValueOf(tc.input)))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectIncludes, includes)
		})
	}

	keyRangeOpenClosed, err := NewKeyRangeBound(safejs.Safe(js.ValueOf(0)), safejs.Safe(js.ValueOf(100)), true, false)
	assert.NoError(t, err)
	for _, tc := range []struct {
		name           string // auto-filled, satisfies paralleltest linter
		input          int
		expectIncludes bool
	}{
		{input: -1, expectIncludes: false},
		{input: 0, expectIncludes: false},
		{input: 50, expectIncludes: true},
		{input: 100, expectIncludes: true},
	} {
		tc.name = fmt.Sprint("open closed ", tc.input)
		tc := tc // keep loop-local copy of test case for parallel runs
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			includes, err := keyRangeOpenClosed.Includes(safejs.Safe(js.ValueOf(tc.input)))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectIncludes, includes)
		})
	}
}

func TestNewKeyRangeLowerBound(t *testing.T) {
	t.Parallel()
	keyRangeOpen, err := NewKeyRangeLowerBound(safejs.Safe(js.ValueOf(0)), true)
	assert.NoError(t, err)
	for _, tc := range []struct {
		input          int
		expectIncludes bool
	}{
		{input: -1, expectIncludes: false},
		{input: 0, expectIncludes: false},
		{input: 100, expectIncludes: true},
	} {
		tc := tc // keep loop-local copy of test case for parallel runs
		t.Run(fmt.Sprint("open ", tc.input), func(t *testing.T) {
			t.Parallel()
			includes, err := keyRangeOpen.Includes(safejs.Safe(js.ValueOf(tc.input)))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectIncludes, includes)
		})
	}

	keyRangeClosed, err := NewKeyRangeLowerBound(safejs.Safe(js.ValueOf(0)), false)
	assert.NoError(t, err)
	for _, tc := range []struct {
		name           string // auto-filled, satisfies paralleltest linter
		input          int
		expectIncludes bool
	}{
		{input: -1, expectIncludes: false},
		{input: 0, expectIncludes: true},
		{input: 100, expectIncludes: true},
	} {
		tc.name = fmt.Sprint("closed ", tc.input)
		tc := tc // keep loop-local copy of test case for parallel runs
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			includes, err := keyRangeClosed.Includes(safejs.Safe(js.ValueOf(tc.input)))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectIncludes, includes)
		})
	}
}

func TestNewKeyRangeUpperBound(t *testing.T) {
	t.Parallel()
	keyRangeOpen, err := NewKeyRangeUpperBound(safejs.Safe(js.ValueOf(100)), true)
	assert.NoError(t, err)
	for _, tc := range []struct {
		name           string // auto-filled, satisfies paralleltest linter
		input          int
		expectIncludes bool
	}{
		{input: 0, expectIncludes: true},
		{input: 100, expectIncludes: false},
		{input: 101, expectIncludes: false},
	} {
		tc := tc // keep loop-local copy of test case for parallel runs
		t.Run(fmt.Sprint("open ", tc.input), func(t *testing.T) {
			t.Parallel()
			includes, err := keyRangeOpen.Includes(safejs.Safe(js.ValueOf(tc.input)))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectIncludes, includes)
		})
	}

	keyRangeClosed, err := NewKeyRangeUpperBound(safejs.Safe(js.ValueOf(100)), false)
	assert.NoError(t, err)
	for _, tc := range []struct {
		name           string // auto-filled, satisfies paralleltest linter
		input          int
		expectIncludes bool
	}{
		{input: 0, expectIncludes: true},
		{input: 100, expectIncludes: true},
		{input: 101, expectIncludes: false},
	} {
		tc.name = fmt.Sprint("closed ", tc.input)
		tc := tc // keep loop-local copy of test case for parallel runs
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			includes, err := keyRangeClosed.Includes(safejs.Safe(js.ValueOf(tc.input)))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectIncludes, includes)
		})
	}
}

func TestNewKeyRangeOnly(t *testing.T) {
	t.Parallel()
	keyRange, err := NewKeyRangeOnly(safejs.Safe(js.ValueOf(100)))
	assert.NoError(t, err)
	for _, tc := range []struct {
		name           string // auto-filled, satisfies paralleltest linter
		input          int
		expectIncludes bool
	}{
		{input: 0, expectIncludes: false},
		{input: 100, expectIncludes: true},
	} {
		tc.name = fmt.Sprint(tc.input)
		tc := tc // keep loop-local copy of test case for parallel runs
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			includes, err := keyRange.Includes(safejs.Safe(js.ValueOf(tc.input)))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectIncludes, includes)
		})
	}
}

func TestKeyRangeBoundProperties(t *testing.T) {
	t.Parallel()
	keyRange, err := NewKeyRangeBound(safejs.Safe(js.ValueOf(0)), safejs.Safe(js.ValueOf(100)), false, true)
	assert.NoError(t, err)

	lower, err := keyRange.Lower()
	assert.NoError(t, err)
	assert.Equal(t, safejs.Safe(js.ValueOf(0)), lower)

	lowerOpen, err := keyRange.LowerOpen()
	assert.NoError(t, err)
	assert.Equal(t, false, lowerOpen)

	upper, err := keyRange.Upper()
	assert.NoError(t, err)
	assert.Equal(t, safejs.Safe(js.ValueOf(100)), upper)

	upperOpen, err := keyRange.UpperOpen()
	assert.NoError(t, err)
	assert.Equal(t, true, upperOpen)
}
