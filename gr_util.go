package gr

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	r "reflect"
	"sync"
	"unsafe"
)

var (
	errUrlAppend = fmt.Errorf(`[gt] failed to append to URL path: unexpected empty string`)
	bytesNewline = []byte("\n")
)

func errResUnexpected(desc string) error {
	return fmt.Errorf(`unexpected %v response`, desc)
}

func errResUnexpectedWithEmptyBody(desc string) error {
	return fmt.Errorf(`unexpected %v response with empty body`, desc)
}

func errResUnexpectedFailedToReadBody(desc string) error {
	return fmt.Errorf(`unexpected %v response; failed to read response body`, desc)
}

func errReqBodyClone(err error) error {
	return fmt.Errorf(`[gr] failed to clone request body: %w`, err)
}

func errForkRead(err error) error {
	return fmt.Errorf(`[gr] failed to read for forking: %w`, err)
}

func errWrite(err error) error {
	return fmt.Errorf(`[gr] failed to write: %w`, err)
}

/*
Allocation-free conversion. Reinterprets a byte slice as a string. Borrowed from
the standard library. Reasonably safe. Should not be used when the underlying
byte array is volatile, for example when it's part of a scratch buffer during
SQL scanning.
*/
func bytesString(input []byte) string {
	return *(*string)(unsafe.Pointer(&input))
}

// Must be deferred.
func rec(ptr *error) {
	val := recover()
	if val == nil {
		return
	}

	err, _ := val.(error)
	if err != nil {
		*ptr = err
		return
	}

	panic(val)
}

func canonKey(val string) string { return textproto.CanonicalMIMEHeaderKey(val) }

func readBody(body io.Reader) []byte {
	val, err := io.ReadAll(body)
	if err != nil {
		panic(fmt.Errorf(`[gr] failed to read response body: %w`, err))
	}
	return val
}

var (
	ctxOffsetLazy sync.Once
	ctxOffsetVal  uintptr
	ctxType       = r.TypeOf((*context.Context)(nil)).Elem()
)

func ctxOffsetFind() {
	typ := r.TypeOf((*http.Request)(nil)).Elem()
	for i := range iter(typ.NumField()) {
		field := typ.Field(i)
		if field.Type == ctxType {
			ctxOffsetVal = field.Offset
			return
		}
	}
	panic(`failed to identify offset of "context.Context" in "http.Request"`)
}

func ctxOffset() uintptr {
	ctxOffsetLazy.Do(ctxOffsetFind)
	return ctxOffsetVal
}

func iter(count int) []struct{} { return make([]struct{}, count) }

func typeOf(typ interface{}) r.Type {
	return typeDeref(r.TypeOf(typ))
}

func typeDeref(typ r.Type) r.Type {
	for typ != nil && typ.Kind() == r.Ptr {
		typ = typ.Elem()
	}
	return typ
}

func valueOf(val interface{}) r.Value {
	return valueDeref(r.ValueOf(val))
}

func valueDeref(val r.Value) r.Value {
	for val.Kind() == r.Ptr {
		if val.IsNil() {
			return r.Value{}
		}
		val = val.Elem()
	}
	return val
}

// Copied from `github.com/mitranim/gax` and tested there.
func growBytes(prev []byte, size int) []byte {
	len, cap := len(prev), cap(prev)
	if cap-len >= size {
		return prev
	}

	next := make([]byte, len, 2*cap+size)
	copy(next, prev)
	return next
}

func appendBodyPreview(buf, body []byte) []byte {
	size := len(body)
	if size == 0 {
		return buf
	}

	const limit = 4096
	buf = append(buf, `; body: `...)

	if len(body) > limit {
		buf = append(buf, body[:limit]...)
		buf = append(buf, ` ... <truncated>`...)
	} else {
		buf = append(buf, body...)
	}

	return buf
}

func cloneStrings(src []string) []string {
	if src == nil {
		return nil
	}
	out := make([]string, len(src))
	copy(out, src)
	return out
}
