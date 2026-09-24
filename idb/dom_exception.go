//go:build js && wasm
// +build js,wasm

package idb

import (
	"fmt"
	"syscall/js"

	"github.com/hack-pad/safejs"
)

var jsDOMException = js.Global().Get("DOMException")

// call invokes method m on v. A thrown DOMException returns as a DOMException,
// so callers can match it by name with errors.Is. safejs.Value.Call returns an
// opaque error that hides the thrown value.
func call(v safejs.Value, m string, args ...any) (result safejs.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = thrownError(r)
		}
	}()
	return safejs.Safe(safejs.Unsafe(v).Call(m, jsArgs(args)...)), nil
}

// jsArgs unwraps safejs values, including values nested in slices and maps,
// into arguments syscall/js accepts.
func jsArgs(args []any) []any {
	values := make([]any, len(args))
	for i, arg := range args {
		values[i] = jsArg(arg)
	}
	return values
}

func jsArg(arg any) any {
	switch arg := arg.(type) {
	case safejs.Value:
		return safejs.Unsafe(arg)
	case []any:
		return jsArgs(arg)
	case map[string]any:
		values := make(map[string]any, len(arg))
		for key, value := range arg {
			values[key] = jsArg(value)
		}
		return values
	default:
		return arg
	}
}

// thrownError converts a value recovered from a panic into an error. A thrown
// DOMException becomes a DOMException. Returns nil for a nil value.
func thrownError(r any) error {
	switch r := r.(type) {
	case nil:
		return nil
	case js.Error:
		if !r.Value.InstanceOf(jsDOMException) {
			return r
		}
		domException, err := parseJSDOMException(safejs.Safe(r.Value))
		if err != nil {
			return r
		}
		return domException
	case js.Value:
		return thrownError(js.Error{Value: r})
	case error:
		return r
	default:
		return fmt.Errorf("%+v", r)
	}
}

func domExceptionAsError(value safejs.Value) error {
	truthy, err := value.Truthy()
	if err != nil || !truthy {
		return err
	}
	domException, err := parseJSDOMException(value)
	if err != nil {
		return err
	}
	return domException
}

// DOMException is a JavaScript DOMException with a standard name.
// Use errors.Is() to compare by name.
type DOMException struct {
	name    string
	message string
}

// NewDOMException returns a new DOMException with the given name.
// Only useful for errors.Is() comparisons with errors returned from idb.
func NewDOMException(name string) DOMException {
	return DOMException{name: name}
}

func parseJSDOMException(value safejs.Value) (DOMException, error) {
	name, err := value.Get("name")
	if err != nil {
		return DOMException{}, err
	}
	nameStr, err := name.String()
	if err != nil {
		return DOMException{}, err
	}
	message, err := value.Get("message")
	if err != nil {
		return DOMException{}, err
	}
	messageStr, err := message.String()
	if err != nil {
		return DOMException{}, err
	}
	return DOMException{
		name:    nameStr,
		message: messageStr,
	}, nil
}

func (e DOMException) Error() string {
	if e.message == "" {
		return e.name
	}
	return e.name + ": " + e.message
}

// Is returns true target is a DOMException and matches this DOMException's name. Use 'errors.Is()' to call it.
func (e DOMException) Is(target error) bool {
	targetDOMException, ok := target.(DOMException)
	return ok && targetDOMException.name == e.name
}
