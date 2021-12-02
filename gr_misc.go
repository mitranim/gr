package gr

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	r "reflect"
	"strconv"
	"strings"
)

const (
	Type      = `Content-Type`
	TypeJson  = `application/json; charset=utf-8`
	TypeForm  = `application/x-www-form-urlencoded; charset=utf-8`
	TypeMulti = `multipart/form-data; charset=utf-8`
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

/*
Short for "transport". Function type that implements `http.RoundTripper` by
calling itself. Can be used as `http.Client.Transport` or `gr.Cli.Transport`.
*/
type Trans func(*http.Request) (*http.Response, error)

// Implement `http.RoundTripper`.
func (self Trans) RoundTrip(req *http.Request) (*http.Response, error) {
	if self != nil {
		return self(req)
	}
	return nil, nil
}

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

/*
Appends the string representation of the input to the path of the given URL,
slash-separated. The input must be string-encodable following the rules of
`gr.Str`, and the resulting string must be non-empty, otherwise this panics to
safeguard against calling external endpoints on the wrong address. Mutates the
given URL and returns it. If the input URL is nil, creates and returns a new
non-nil instance. For correctness, you must reassign the output instead of
relying on mutation.
*/
func UrlAppend(ref *url.URL, val interface{}) *url.URL {
	str := Str(val)
	if str == `` {
		panic(errUrlAppend)
	}

	if ref == nil {
		return &url.URL{Path: str}
	}

	ref.Path = path.Join(ref.Path, str)
	return ref
}

/*
Appends the string representations of the input values to the path of the given
URL, slash-separated. Uses `gr.UrlAppend` for each segment. Each segment must
be non-empty, otherwise this panics to safeguard against calling external
endpoints on the wrong address. Mutates the given URL and returns it. If the
input URL is nil, creates and returns a new non-nil instance. For correctness,
you must reassign the output instead of relying on mutation.
*/
func UrlJoin(ref *url.URL, vals ...interface{}) *url.URL {
	for _, val := range vals {
		ref = UrlAppend(ref, val)
	}
	return ref
}

/*
Missing feature of the standard library: return a string representation of a
primitive or "intentionally" stringable value, without allowing arbitrary
non-stringable inputs. Differences from `fmt.Sprint`:

	* Allow ONLY the following inputs:

		* nil                       -> return ""
		* `fmt.Stringer`            -> call `.String()`
		* Built-in primitive types  -> use "strconv"

	* Automatically dereference pointers to supported types.
	  Nil pointers are considered equivalent to nil.

	* Panic for all other inputs.

	* Don't swallow encoding errors.
*/
func Str(src interface{}) string {
	if src == nil {
		return ``
	}

	impl, _ := src.(fmt.Stringer)
	if impl != nil {
		return impl.String()
	}

	typ := typeOf(src)
	val := valueOf(src)

	switch typ.Kind() {
	case r.Int8, r.Int16, r.Int32, r.Int64, r.Int:
		if val.IsValid() {
			return strconv.FormatInt(val.Int(), 10)
		}
		return ``

	case r.Uint8, r.Uint16, r.Uint32, r.Uint64, r.Uint:
		if val.IsValid() {
			return strconv.FormatUint(val.Uint(), 10)
		}
		return ``

	case r.Float32, r.Float64:
		if val.IsValid() {
			return strconv.FormatFloat(val.Float(), 'f', -1, 64)
		}
		return ``

	case r.Bool:
		if val.IsValid() {
			return strconv.FormatBool(val.Bool())
		}
		return ``

	case r.String:
		if val.IsValid() {
			return val.String()
		}
		return ``

	default:
		/**
		Doesn't report the value itself because it could be arbitrarily large, not
		encodable without another panic, or include information that shouldn't be
		exposed.
		*/
		panic(fmt.Errorf(
			`[gt] failed to encode value of unsupported type %q of kind %q as string`,
			typ, typ.Kind(),
		))
	}
}
