package gr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"unsafe"
)

/*
Returns a new request with the given context.
Shortcut for `new(gr.Req).Ctx(ctx)`.
*/
func Ctx(ctx context.Context) *Req { return new(Req).Ctx(ctx) }

/*
Returns a new request with the given client.
Shortcut for `new(gr.Req).Cli(val)`.
The name is short for "client", not "CLI".
*/
func Cli(val *http.Client) *Req { return new(Req).Cli(val) }

/*
Returns a new request with the given URL string.
Shortcut for `new(gr.Req).To(val)`.
*/
func To(val string) *Req { return new(Req).To(val) }

/*
Returns a new request with the given URL.
Shortcut for `new(gr.Req).Url(val)`.
*/
func Url(val *url.URL) *Req { return new(Req).Url(val) }

/*
Returns a new request with the given URL path.
Shortcut for `new(gr.Req).Path(val, vals...)`.
*/
func Path(val string, vals ...interface{}) *Req {
	return new(Req).Path(val, vals...)
}

/*
Returns a new request with the given method.
Shortcut for `new(gr.Req).Meth(val)`.
*/
func Meth(val string) *Req { return new(Req).Meth(val) }

// Returns a new "GET" request. Shortcut for `new(gr.Req).Get()`.
func Get() *Req { return new(Req).Get() }

// Returns a new "POST" request. Shortcut for `new(gr.Req).Post()`.
func Post() *Req { return new(Req).Post() }

// Returns a new "PUT" request. Shortcut for `new(gr.Req).Put()`.
func Put() *Req { return new(Req).Put() }

// Returns a new "PATCH" request. Shortcut for `new(gr.Req).Patch()`.
func Patch() *Req { return new(Req).Patch() }

// Returns a new "DELETE" request. Shortcut for `new(gr.Req).Delete()`.
func Delete() *Req { return new(Req).Delete() }

// Returns a new "OPTIONS" request. Shortcut for `new(gr.Req).Options()`.
func Options() *Req { return new(Req).Options() }

/*
Returns a new request with pre-initialized non-zero context, method, URL, and
header. Shortcut for `new(gr.Req).Init()`.
*/
func Init() *Req { return new(Req).Init() }

/*
Alias of `http.Request` with a fluent builder-style API. Freely castable to and
from `http.Request`. All methods are defined on `*gr.Req` and may mutate the
receiver or its inner references. To store and copy a "partially built"
request, use `(*gr.Req).Clone`, but beware that it doesn't copy the body.
*/
type Req http.Request

// Free cast to `*http.Request`.
func (self *Req) Req() *http.Request { return (*http.Request)(self) }

/*
Initializes context, `.Method`, `.URL`, `.Header` to non-zero values, similar to
how `http.NewRequest` would have done it. Mutates and returns the receiver.
*/
func (self *Req) Init() *Req {
	return self.initCtx().initMeth().initUrl().initHead()
}

/*
Sets the inner context to the exact given value, without nil checks or
fallbacks. Mutates and returns the receiver.
*/
func (self *Req) Ctx(ctx context.Context) *Req {
	if self == nil {
		return nil
	}
	*self.ctx() = ctx
	return self
}

/*
Returns the inner context as-is. Like `(*http.Request).Context`, but without the
hidden fallback on `context.Background`.
*/
func (self *Req) Context() context.Context {
	if self == nil {
		return nil
	}
	return *self.ctx()
}

func (self *Req) ctx() *context.Context {
	return (*context.Context)(unsafe.Pointer(uintptr(unsafe.Pointer(self)) + ctxOffset()))
}

/*
Sets the given HTTP client, unsafely reusing the `.TLS` field which is normally
unused in client requests. Passing nil clears the field. The client is
automatically used by `(*gr.Req).Res` and `(*gr.Req).ResCatch`. The name is
short for "client", not "CLI".
*/
func (self *Req) Cli(val *http.Client) *Req {
	*self.cli() = val
	return self
}

/*
Returns the HTTP client previously set by `(*gr.Req).Cli`, unsafely reusing the
`.TLS` field which is normally unused in client requests. The default is nil.
*/
func (self *Req) Client() *http.Client { return *self.cli() }

func (self *Req) cli() **http.Client {
	return (**http.Client)(unsafe.Pointer(&self.TLS))
}

// True if `.Method` is "", "GET", "HEAD" or "OPTIONS".
func (self *Req) IsReadOnly() bool { return IsReadOnly(self.Method) }

// Sets `.Method` to the given value. Mutates and returns the receiver.
func (self *Req) Meth(val string) *Req {
	self.Method = val
	return self
}

// Shortcut for `self.Meth(http.MethodGet)`.
func (self *Req) Get() *Req { return self.Meth(http.MethodGet) }

// Shortcut for `self.Meth(http.MethodPost)`.
func (self *Req) Post() *Req { return self.Meth(http.MethodPost) }

// Shortcut for `self.Meth(http.MethodPut)`.
func (self *Req) Put() *Req { return self.Meth(http.MethodPut) }

// Shortcut for `self.Meth(http.MethodPatch)`.
func (self *Req) Patch() *Req { return self.Meth(http.MethodPatch) }

// Shortcut for `self.Meth(http.MethodDelete)`.
func (self *Req) Delete() *Req { return self.Meth(http.MethodDelete) }

// Shortcut for `self.Meth(http.MethodOptions)`.
func (self *Req) Options() *Req { return self.Meth(http.MethodOptions) }

/*
Parses the input via `url.Parse` and sets `.URL` to the result. Panics on
parsing errors. Mutates and returns the receiver.
*/
func (self *Req) To(src string) *Req {
	val, err := url.Parse(src)
	if err != nil {
		panic(fmt.Errorf(`[gr] failed to parse request destination: %w`, err))
	}
	return self.Url(val)
}

// Sets the given value as `.URL`. Mutates and returns the receiver.
func (self *Req) Url(val *url.URL) *Req {
	self.URL = val
	return self
}

/*
Sets the given value as `.URL.Path`, creating a new URL reference if the URL was
nil. Mutates and returns the receiver.
*/
func (self *Req) Path(val string, vals ...interface{}) *Req {
	self = self.initUrl()
	self.URL.Path = val
	self.URL = UrlJoin(self.URL, vals...)
	return self
}

/*
Uses `gr.UrlAppend` to append the input to the URL path, slash-separated.
Mutates and returns the receiver.
*/
func (self *Req) Append(val interface{}) *Req {
	self = self.initUrl()
	self.URL = UrlAppend(self.URL, val)
	return self
}

/*
Uses `gr.UrlJoin` to append the inputs to the URL path, slash-separated. Mutates
and returns the receiver.
*/
func (self *Req) Join(vals ...interface{}) *Req {
	self = self.initUrl()
	self.URL = UrlJoin(self.URL, vals...)
	return self
}

/*
Sets the given value as `.URL.RawQuery`, creating a new URL reference if the URL
was nil. Mutates and returns the receiver.
*/
func (self *Req) RawQuery(val string) *Req {
	self = self.initUrl()
	self.URL.RawQuery = val
	return self
}

/*
Shortcut for `self.RawQuery(url.Values(val).Encode())`. Accepts an "anonymous"
type because all alias types such as `url.Values` are automatically castable
into it.
*/
func (self *Req) Query(val map[string][]string) *Req {
	return self.RawQuery(url.Values(val).Encode())
}

/*
Sets the given value as `.Header`. Mutates and returns the receiver. Accepts
an "anonymous" type because all alias types such as `http.Header` and `gr.Head`
are automatically castable into it.
*/
func (self *Req) Head(val map[string][]string) *Req {
	self.Header = val
	return self
}

/*
Deletes the given entry in `.Header` by using `gr.Head.Del`. May mutate
`.Header`, but not the slices contained therein. Mutates and returns the
receiver.
*/
func (self *Req) HeadDel(key string) *Req {
	self.Header = Head(self.Header).Del(key).Header()
	return self
}

/*
Appends the given key-value to `.Header` by using `gr.Head.Add`. Allocates the
header if necessary. May mutate `.Header` and an existing slice corresponding
to the key. Mutates and returns the receiver.
*/
func (self *Req) HeadAdd(key, val string) *Req {
	self.Header = Head(self.Header).Add(key, val).Header()
	return self
}

/*
Sets the given key-value in `.Header` by using `gr.Head.Set`. Allocates the
header if necessary. May mutate `.Header`, but not the slices contained
therein. Mutates and returns the receiver.
*/
func (self *Req) HeadSet(key, val string) *Req {
	self.Header = Head(self.Header).Set(key, val).Header()
	return self
}

/*
Replaces the given key-values entry in `.Header` by using `gr.Head.Replace`.
Allocates the header if necessary. May mutate `.Header`, but not the slices
contained therein. Mutates and returns the receiver.
*/
func (self *Req) HeadReplace(key string, vals ...string) *Req {
	self.Header = Head(self.Header).Replace(key, vals...).Header()
	return self
}

/*
Patches the header by using `gr.Head.Patch`. Allocates the header if necessary.
May mutate `.Header`, but not the slices contained therein. Mutates and returns
the receiver. Accepts an "anonymous" type because all alias types such as
`http.Header` and `gr.Head` are automatically castable into it.
*/
func (self *Req) HeadPatch(head map[string][]string) *Req {
	self.Header = Head(self.Header).Patch(head).Header()
	return self
}

/*
Shortcut for setting the "Content-Type" header. If the input is "", removes the
header instead. Mutates and returns the receiver.
*/
func (self *Req) Type(typ string) *Req {
	if typ == `` {
		return self.HeadDel(Type)
	}
	return self.HeadSet(Type, typ)
}

/*
Shortcut for setting the "Content-Type: application/json" header. Mutates and
returns the receiver.
*/
func (self *Req) TypeJson() *Req { return self.Type(TypeJson) }

/*
Shortcut for setting the "Content-Type: application/x-www-form-urlencoded"
header. Mutates and returns the receiver.
*/
func (self *Req) TypeForm() *Req { return self.Type(TypeForm) }

/*
Shortcut for setting the "Content-Type: multipart/form-data" header. Mutates and
returns the receiver.
*/
func (self *Req) TypeMulti() *Req { return self.Type(TypeMulti) }

/*
Uses the given string as the request body, updating the following fields:

	* `.ContentLength` -> input length, in bytes rather than characters.
	* `.Body`          -> nil or `gr.NewStringReadCloser` from input.
	* `.GetBody`       -> nil or function returning `gr.NewStringReadCloser` from input.

If the input is empty, the listed fields are set to zero values, otherwise the
fields are set to non-zero values. Mutates and returns the receiver.
*/
func (self *Req) String(val string) *Req {
	self.ContentLength = int64(len(val))

	if self.ContentLength == 0 {
		self.GetBody = nil
		self.Body = nil
		return self
	}

	self.GetBody = func() (io.ReadCloser, error) { return NewStringReadCloser(val), nil }
	self.Body = NewStringReadCloser(val)
	return self
}

/*
Uses the given chunk as the request body, updating the following fields:

	* `.ContentLength` -> input length, in bytes rather than characters.
	* `.Body`          -> nil or `gr.NewBytesReadCloser` from input.
	* `.GetBody`       -> nil or function returning `gr.NewBytesReadCloser` from input.

If the input is empty, the listed fields are set to zero values, otherwise the
fields are set to non-zero values. Mutates and returns the receiver.
*/
func (self *Req) Bytes(val []byte) *Req {
	self.ContentLength = int64(len(val))

	if self.ContentLength == 0 {
		self.GetBody = nil
		self.Body = nil
		return self
	}

	self.GetBody = func() (io.ReadCloser, error) { return NewBytesReadCloser(val), nil }
	self.Body = NewBytesReadCloser(val)
	return self
}

/*
URL-encodes the given vals as the request body. Shortcut for
`self.String(url.Values(val).Encode())`. Also sets `.ContentLength` and
`.GetBody`. Accepts an "anonymous" type because all alias types such as
`url.Values` are automatically castable into it. Mutates and returns the
receiver.
*/
func (self *Req) Vals(val map[string][]string) *Req {
	return self.String(url.Values(val).Encode())
}

/*
URL-encodes the given vals as the request body. Also sets the header
"Content-Type: application/x-www-form-urlencoded", as well as fields
`.ContentLength` and `.GetBody`. Shortcut for `self.TypeForm().Vals(val)`.
Accepts an "anonymous" type because all alias types such as `url.Values` are
automatically castable into it. Mutates and returns the receiver.
*/
func (self *Req) FormVals(val map[string][]string) *Req {
	return self.TypeForm().Vals(val)
}

/*
JSON-encodes an arbitrary value, using it as the request body. Also sets the
header "Content-Type: application/json", as well as fields `.ContentLength` and
`.GetBody`. Panics if JSON encoding fails. Use `(*gr.Req).JsonCatch` to catch
those panics. Mutates and returns the receiver.
*/
func (self *Req) Json(val interface{}) *Req {
	self = self.TypeJson()

	/**
	Questionable but convenient special case. Allows the calling code to use this
	unconditionally, regardless of the resulting HTTP method, without having
	issues with clients that reject GET requests with a non-empty body. Also
	avoids wasting performance in this very common case.
	*/
	if self.IsReadOnly() && val == nil {
		return self.ReadCloser(nil)
	}

	chunk, err := json.Marshal(val)
	if err != nil {
		panic(fmt.Errorf(`[gr] failed to JSON-encode request body: %w`, err))
	}
	return self.Bytes(chunk)
}

/*
Same as `(*gr.Req).Json`, but if JSON encoding fails, returns an error instead
of panicking.
*/
func (self *Req) JsonCatch(val interface{}) (err error) {
	defer rec(&err)
	self.Json(val)
	return
}

/*
Assumes that the given string is valid JSON, and uses it as the request body.
Also sets "Content-Type: application/json". Shortcut for
`self.TypeJson().String(val)`. Mutates and returns the receiver.
*/
func (self *Req) JsonString(val string) *Req {
	return self.TypeJson().String(val)
}

/*
Assumes that the given chunk is valid JSON, and uses it as the request body.
Also sets "Content-Type: application/json". Shortcut for
`self.TypeJson().Bytes(val)`. Mutates and returns the receiver.
*/
func (self *Req) JsonBytes(val []byte) *Req {
	return self.TypeJson().Bytes(val)
}

/*
Shortcut for setting `.Body` and returning the request. Sets the body as-is
without affecting other fields. Mutates and returns the receiver.
*/
func (self *Req) ReadCloser(val io.ReadCloser) *Req {
	self.Body = val
	return self
}

/*
Sets the given reader as the request body. If the reader is nil, sets nil.
Otherwise wraps it in `io.NopCloser`. Mutates and returns the receiver.
*/
func (self *Req) Reader(val io.Reader) *Req {
	if val == nil {
		return self.ReadCloser(nil)
	}
	return self.ReadCloser(io.NopCloser(val))
}

/*
Returns a deep copy, like `(*http.Request).Clone`, but without forcing you to
provide a context. Cloning allows to reuse partially built requests, like
templates. This preserves everything, including the previous context and
client. Inner mutable references such as `.URL` and `.Header` are deeply
cloned. However, `.Body` is not cloned, and may not be reusable.
*/
func (self *Req) Clone() *Req {
	return (*Req)(self.Req().Clone(self.Req().Context())).Ctx(self.Context())
}

/*
Short for "response". Shortcut for `(*gr.Res).CliRes(self.Client())`, which uses
`http.DefaultClient` if no client was given. Returns the response as `*gr.Res`.
Panics on transport errors, but NOT in case of successful HTTP responses with
non-OK HTTP status codes. To avoid panics, use `(*gr.Req).ResCatch`.

The caller MUST close the response body by calling `*gr.Res.Done` or its other
reading or closing methods.
*/
func (self *Req) Res() *Res {
	return self.CliRes(self.Client())
}

/*
Variant of `(*gr.Req).Res` that returns an error instead of panicking. If the
response is non-nil, the caller MUST close the response body by calling
`*gr.Res.Done` or its other reading or closing methods.
*/
func (self *Req) ResCatch() (_ *Res, err error) {
	defer rec(&err)
	return self.Res(), nil
}

/*
Short for "client response" or "response using client". Performs the request
using the given client, returning the response as `*gr.Res`. If the client is
nil, uses `http.DefaultClient`. Panics on transport errors, but NOT in case of
successful HTTP responses with non-OK HTTP status codes. To avoid panics, use
`(*gr.Req).CliResCatch`.

The caller MUST close the response body by calling `*gr.Res.Done` or its other
reading or closing methods.
*/
func (self *Req) CliRes(cli *http.Client) *Res {
	if cli == nil {
		cli = http.DefaultClient
	}

	res, err := cli.Do(self.Init().Req())
	if err != nil {
		panic(fmt.Errorf(`[gr] failed to perform HTTP request: %w`, err))
	}

	return (*Res)(res)
}

/*
Variant of `(*gr.Req).CliRes` that returns an error instead of panicking. If the
response is non-nil, the caller MUST close the response body by calling
`*gr.Res.Done` or its other reading or closing methods.
*/
func (self *Req) CliResCatch(cli *http.Client) (_ *Res, err error) {
	defer rec(&err)
	return self.CliRes(cli), nil
}

func (self *Req) initCtx() *Req {
	if self.Context() == nil {
		return self.Ctx(context.Background())
	}
	return self
}

func (self *Req) initMeth() *Req {
	if self.Method == `` {
		self.Method = http.MethodGet
	}
	return self
}

func (self *Req) initUrl() *Req {
	if self.URL == nil {
		self.URL = &url.URL{}
	}
	return self
}

func (self *Req) initHead() *Req {
	if self.Header == nil {
		self.Header = http.Header{}
	}
	return self
}
