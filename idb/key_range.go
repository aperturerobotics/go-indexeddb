//go:build js && wasm
// +build js,wasm

package idb

import (
	"github.com/hack-pad/safejs"
)

func GetJsIDBKeyRange() (safejs.Value, error) {
	return safejs.Global().Get("IDBKeyRange")
}

// KeyRange represents a continuous interval over some data type that is used for keys. Records can be retrieved from ObjectStore and Index objects using keys or a range of keys.
type KeyRange struct {
	jsKeyRange safejs.Value
}

// WrapKeyRange wraps a safejs.Value into a KeyRange.
func WrapKeyRange(jsKeyRange safejs.Value) *KeyRange {
	return &KeyRange{jsKeyRange}
}

// NewKeyRangeBound creates a new key range with the specified upper and lower bounds.
// The bounds can be open (that is, the bounds exclude the endpoint values) or closed (that is, the bounds include the endpoint values).
func NewKeyRangeBound(lower, upper safejs.Value, lowerOpen, upperOpen bool) (*KeyRange, error) {
	jsIDBKeyRange, err := GetJsIDBKeyRange()
	if err != nil {
		return nil, err
	}
	keyRange, err := jsIDBKeyRange.Call("bound", lower, upper, lowerOpen, upperOpen)
	if err != nil {
		return nil, tryAsDOMException(err)
	}
	return WrapKeyRange(keyRange), nil
}

// NewKeyRangeLowerBound creates a new key range with only a lower bound.
func NewKeyRangeLowerBound(lower safejs.Value, open bool) (*KeyRange, error) {
	jsIDBKeyRange, err := GetJsIDBKeyRange()
	if err != nil {
		return nil, err
	}
	keyRange, err := jsIDBKeyRange.Call("lowerBound", lower, open)
	if err != nil {
		return nil, tryAsDOMException(err)
	}
	return WrapKeyRange(keyRange), nil
}

// NewKeyRangeUpperBound creates a new key range with only an upper bound.
func NewKeyRangeUpperBound(upper safejs.Value, open bool) (*KeyRange, error) {
	jsIDBKeyRange, err := GetJsIDBKeyRange()
	if err != nil {
		return nil, err
	}
	keyRange, err := jsIDBKeyRange.Call("upperBound", upper, open)
	if err != nil {
		return nil, tryAsDOMException(err)
	}
	return WrapKeyRange(keyRange), nil
}

// NewKeyRangeOnly creates a new key range containing a single value.
func NewKeyRangeOnly(only safejs.Value) (*KeyRange, error) {
	jsIDBKeyRange, err := GetJsIDBKeyRange()
	if err != nil {
		return nil, err
	}
	keyRange, err := jsIDBKeyRange.Call("only", only)
	if err != nil {
		return nil, tryAsDOMException(err)
	}
	return WrapKeyRange(keyRange), nil
}

// Lower returns the lower bound of the key range.
func (k *KeyRange) Lower() (safejs.Value, error) {
	lower, err := k.jsKeyRange.Get("lower")
	return lower, err
}

// Upper returns the upper bound of the key range.
func (k *KeyRange) Upper() (safejs.Value, error) {
	upper, err := k.jsKeyRange.Get("upper")
	return upper, err
}

// LowerOpen returns false if the lower-bound value is included in the key range.
func (k *KeyRange) LowerOpen() (bool, error) {
	lowerOpen, err := k.jsKeyRange.Get("lowerOpen")
	if err != nil {
		return false, err
	}
	return lowerOpen.Bool()
}

// UpperOpen returns false if the upper-bound value is included in the key range.
func (k *KeyRange) UpperOpen() (bool, error) {
	upperOpen, err := k.jsKeyRange.Get("upperOpen")
	if err != nil {
		return false, err
	}
	return upperOpen.Bool()
}

// Includes returns a boolean indicating whether a specified key is inside the key range.
func (k *KeyRange) Includes(key safejs.Value) (bool, error) {
	includes, err := k.jsKeyRange.Call("includes", key)
	if err != nil {
		return false, tryAsDOMException(err)
	}
	return includes.Bool()
}

// Unwrap unwraps the key range into a safejs.Value.
func (k *KeyRange) Unwrap() safejs.Value {
	return k.jsKeyRange
}
