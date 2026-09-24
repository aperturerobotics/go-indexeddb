//go:build js && wasm
// +build js,wasm

package idb

import (
	"testing"

	"github.com/aperturerobotics/go-indexeddb/idb/internal/assert"
	"github.com/hack-pad/safejs"
)

var domException safejs.Value

func init() {
	var err error
	domException, err = safejs.Global().Get("DOMException")
	if err != nil {
		panic(err)
	}
}

func TestCallReturnsDOMException(t *testing.T) {
	t.Parallel()
	thrower, err := safejs.Global().Get("Function")
	assert.NoError(t, err)
	makeThrower, err := thrower.New(`return { run() { throw new DOMException("message", "name") } }`)
	assert.NoError(t, err)
	obj, err := makeThrower.Invoke()
	assert.NoError(t, err)

	_, err = call(obj, "run")
	assert.Equal(t, DOMException{
		name:    "name",
		message: "message",
	}, err)
}

func TestDOMExceptionAsError(t *testing.T) {
	t.Parallel()
	exceptionJS, err := domException.New("message", "name")
	assert.NoError(t, err)
	exception := domExceptionAsError(exceptionJS)
	assert.Equal(t, DOMException{
		name:    "name",
		message: "message",
	}, exception)

	assert.Equal(t, "name: message", exception.Error())

	assert.ErrorIs(t, exception, DOMException{name: "name"})
	assert.NotErrorIs(t, exception, DOMException{name: "other name"})
}
