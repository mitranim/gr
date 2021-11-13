package gr

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

const (
	Type      = `Content-Type`
	TypeJson  = `application/json`
	TypeForm  = `application/x-www-form-urlencoded`
	TypeMulti = `multipart/form-data`
)

/*
Equivalent to `io.NopCloser(strings.NewReader(val))`, but avoids 1 indirection
and heap allocation by returning `*gr.StringReadCloser` which implements
nop `io.Closer` by itself. Used internally by `(*gr.Req).String`.
*/
func NewStringReadCloser(val string) *StringReadCloser {
	var buf strings.Reader
	buf.Reset(val)
	return &StringReadCloser{buf}
}

// Variant of `strings.Reader` that also implements nop `io.Closer`.
// Used internally by `(*gr.Req).String`.
type StringReadCloser struct{ strings.Reader }

var _ = io.ReadCloser((*StringReadCloser)(nil))

// Nop `io.Closer`.
func (self *StringReadCloser) Close() error { return nil }

/*
Equivalent to `io.NopCloser(bytes.NewReader(val))`, but avoids 1 indirection
and heap allocation by returning `*gr.BytesReadCloser` which implements
nop `io.Closer` by itself. Used internally by `(*gr.Req).Bytes`.
*/
func NewBytesReadCloser(val []byte) *BytesReadCloser {
	var buf bytes.Reader
	buf.Reset(val)
	return &BytesReadCloser{buf}
}

// Variant of `bytes.Reader` that also implements nop `io.Closer`.
// Used internally by `(*gr.Req).Bytes`.
type BytesReadCloser struct{ bytes.Reader }

var _ = io.ReadCloser((*BytesReadCloser)(nil))

// Nop `io.Closer`.
func (self *BytesReadCloser) Close() error { return nil }

// True if given HTTP method is "", "GET", "HEAD" or "OPTIONS".
func IsReadOnly(val string) bool {
	return val == `` ||
		val == http.MethodGet ||
		val == http.MethodHead ||
		val == http.MethodOptions
}

// True if given HTTP status code is between 100 and 199 inclusive.
func IsInfo(val int) bool { return val >= 100 && val <= 199 }

// True if given HTTP status code is between 200 and 299 inclusive.
func IsOk(val int) bool { return val >= 200 && val <= 299 }

// True if given HTTP status code is between 300 and 399 inclusive.
func IsRedir(val int) bool { return val >= 300 && val <= 399 }

// True if given HTTP status code is between 400 and 499 inclusive.
func IsClientErr(val int) bool { return val >= 400 && val <= 499 }

// True if given HTTP status code is between 500 and 599 inclusive.
func IsServerErr(val int) bool { return val >= 500 && val <= 599 }
