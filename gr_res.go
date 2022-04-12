package gr

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
)

/*
Alias of `http.Response` with many shortcuts for inspecting and decoding the
response. Freely castable to and from `http.Response`.

When using `gr`, a response can be obtained by calling `(*gr.Req).Res` or
`(*gr.Req).CliRes`. However, this is also usable for responses obtained from
other sources. You can cast any `*http.Response` to `*gr.Res` to simplify its
inspection and decoding.

The caller MUST close the response body by calling `(*gr.Res).Done` or its other
reading or closing methods.
*/
type Res http.Response

// Free cast to `*http.Response`.
func (self *Res) Res() *http.Response { return (*http.Response)(self) }

// True if response status code is between 100 and 199 inclusive.
func (self *Res) IsInfo() bool { return IsInfo(self.StatusCode) }

// True if response status code is between 200 and 299 inclusive.
func (self *Res) IsOk() bool { return IsOk(self.StatusCode) }

// True if response status code is between 300 and 399 inclusive.
func (self *Res) IsRedir() bool { return IsRedir(self.StatusCode) }

// True if response status code is between 400 and 499 inclusive.
func (self *Res) IsClientErr() bool { return IsClientErr(self.StatusCode) }

// True if response status code is between 500 and 599 inclusive.
func (self *Res) IsServerErr() bool { return IsServerErr(self.StatusCode) }

/*
Asserts that the response is "ok". If not "ok", panics with an error that
includes the response body as text, and closes the body. If "ok", returns the
response unchanged without closing the body.
*/
func (self *Res) Ok() *Res {
	if self.IsOk() {
		return self
	}

	defer self.Done()
	panic(self.Err(`non-OK`))
}

/*
Non-panicking version of `(*gr.Res).Ok`. If not "ok", returns an error that
includes the response body as text, and closes the body. Otherwise returns nil
without closing the body.
*/
func (self *Res) OkCatch() (err error) {
	defer rec(&err)
	self.Ok()
	return
}

/*
Asserts that the response is a redirect. If not a redirect, panics with an error
that includes the response body as text, and closes the body. If a redirect,
returns the response unchanged without closing the body.
*/
func (self *Res) Redir() *Res {
	if self.IsRedir() {
		return self
	}

	defer self.Done()
	panic(self.Err(`non-redirect`))
}

/*
Non-panicking version of `(*gr.Res).Redir`. If not a redirect, returns an error
that includes the response body as text, and closes the body. Otherwise returns
nil without closing the body.
*/
func (self *Res) RedirCatch() (err error) {
	defer rec(&err)
	self.Redir()
	return
}

/*
Short for "location". Returns the response header "Location". Sometimes useful
when inspecting redirect responses.
*/
func (self *Res) Loc() string {
	return self.Header.Get(`Location`)
}

/*
Short for "location URL". Parses the response header "Location" and returns it
as URL. If parsing fails, panics with a descriptive error.
*/
func (self *Res) LocUrl() *url.URL {
	loc := self.Loc()
	val, err := url.Parse(loc)
	if err != nil {
		panic(fmt.Errorf(`[gr] failed to parse redirect location %q: %w`, loc, err))
	}
	return val
}

/*
Non-panicking version of `(*gr.Res).LocUrl`. Returns a parsed redirect location
or a descriptive parse error.
*/
func (self *Res) LocUrlCatch() (_ *url.URL, err error) {
	defer rec(&err)
	return self.LocUrl(), nil
}

/*
Closes the response body if possible, returning the same response. Can be
deferred or used in method chains.
*/
func (self *Res) Done() *Res {
	body := self.Body
	if body != nil {
		_ = body.Close()
	}
	return self
}

// Closes the response body if possible, returning a close error if any.
func (self *Res) CloseErr() error {
	body := self.Body
	if body != nil {
		return body.Close()
	}
	return nil
}

/*
Returns the `Content-Type` header. Note that the content-type may contain
additional parameters, and needs to be parsed before comparing it to a "pure"
media type such as `gr.TypeJson`.
*/
func (self *Res) Type() string {
	if self != nil && self.Header != nil {
		return self.Header.Get(Type)
	}
	return ``
}

/*
Parses the `Content-Type` header via `mime.ParseMediaType`.
Panics on error.
*/
func (self *Res) Media() (_ string, _ map[string]string) {
	src := self.Type()
	if src == `` {
		return
	}

	typ, par, err := mime.ParseMediaType(src)
	if err != nil {
		panic(fmt.Errorf(`[gr] failed to parse media type %q: %w`, src, err))
	}
	return typ, par
}

// Returns the media type parsed via `(*gr.Res).Media`.
func (self *Res) MediaType() (_ string) {
	val, _ := self.Media()
	return val
}

// True if the parsed media type of `Content-Type` is `gr.TypeJson`.
func (self *Res) IsJson() bool { return self.MediaType() == TypeJson }

// True if the parsed media type of `Content-Type` is `gr.TypeForm`.
func (self *Res) IsForm() bool { return self.MediaType() == TypeForm }

// True if the parsed media type of `Content-Type` is `gr.TypeMulti`.
func (self *Res) IsMulti() bool { return self.MediaType() == TypeMulti }

/*
Uses `io.ReadAll` to read the entire response body, returning the resulting
chunk. Panics if reading can't be completed. Always closes the body.
*/
func (self *Res) ReadBytes() []byte {
	body := self.Body
	if body == nil {
		return nil
	}
	defer body.Close()
	return readBody(body)
}

/*
Non-panicking version of `(*gr.Res).ReadBytes`. If body reading fails, returns
an error. Always closes the body.
*/
func (self *Res) ReadBytesCatch() (_ []byte, err error) {
	defer rec(&err)
	return self.ReadBytes(), nil
}

/*
Similar to `(*gr.Res).ReadBytes`, but returns the entire response body as a
string, without copying. Always closes the body.
*/
func (self *Res) ReadString() string { return bytesString(self.ReadBytes()) }

/*
Non-panicking version of `(*gr.Res).ReadString`. If body reading fails, returns
an error. Always closes the body.
*/
func (self *Res) ReadStringCatch() (_ string, err error) {
	defer rec(&err)
	return self.ReadString(), nil
}

/*
Downloads the response body and parses it as URL-encoded/form-encoded content,
using `url.ParseQuery`. Returns the parsed result. Panics if body can't be
read, or if parsing fails. Always closes the body.
*/
func (self *Res) Form() url.Values {
	if self.Body == nil {
		return nil
	}

	val, err := url.ParseQuery(self.ReadString())
	if err != nil {
		panic(fmt.Errorf(`[gr] failed to form-decode response body: %w`, err))
	}

	return val
}

/*
Non-panicking version of `(*gr.Res).Form`. If body can't be read, or if parsing
fails, returns an error. Always closes the body.
*/
func (self *Res) FormCatch() (_ url.Values, err error) {
	defer rec(&err)
	return self.Form(), nil
}

/*
Parses the response body into the given output, which must be either nil or a
pointer. Uses `json.Decoder` to decode from a stream, without buffering the
entire body. Panics on errors. If the output is nil, skips downloading or
decoding. Returns the same response. Always closes the body.
*/
func (self *Res) Json(out interface{}) *Res {
	body := self.Body
	if body == nil {
		return self
	}
	defer body.Close()

	if out == nil {
		return self
	}

	err := json.NewDecoder(body).Decode(out)
	if err != nil {
		panic(fmt.Errorf(`[gr] failed to JSON-decode response body: %w`, err))
	}
	return self
}

/*
Non-panicking version of `(*gr.Res).Json`. Returns an error if body downloading
or parsing fails. Always closes the body.
*/
func (self *Res) JsonCatch(out interface{}) (err error) {
	defer rec(&err)
	self.Json(out)
	return
}

/*
If the response is "ok", decodes the response body into `outVal`. Otherwise
decodes the response body into `outErr`. Both outputs are optional; if the
relevant output is nil, skips downloading and decoding. Returns true if the
response is "ok", and false if not. Panics on downloading or decoding errors.
Always closes the body.
*/
func (self *Res) JsonEither(outVal, outErr interface{}) bool {
	defer self.Done()
	if self.IsOk() {
		self.Json(outVal)
		return true
	}
	self.Json(outErr)
	return false
}

/*
Non-panicking version of `(*gr.Res).JsonEither`. If body can't be read, or if
parsing fails, returns an error. Always closes the body.
*/
func (self *Res) JsonEitherCatch(outVal, outErr interface{}) (_ bool, err error) {
	defer rec(&err)
	return self.JsonEither(outVal, outErr), nil
}

/*
Shortcut for `(*gr.Res).XmlWith` with a nil func. Decodes the response body
without any special XML decoder options.
*/
func (self *Res) Xml(out interface{}) *Res {
	return self.XmlWith(out, nil)
}

/*
Parses the response body into the given output, which must be either nil or a
pointer. Uses `xml.Decoder` to decode from a stream, without buffering the
entire body. The given function is used to customize the decoder, and may be
nil. Panics on errors. If the output is nil, skips downloading or decoding.
Returns the same response. Always closes the body.
*/
func (self *Res) XmlWith(out interface{}, fun func(*xml.Decoder)) *Res {
	body := self.Body
	if body == nil {
		return self
	}
	defer body.Close()

	if out == nil {
		return self
	}

	dec := xml.NewDecoder(body)
	if fun != nil {
		fun(dec)
	}

	err := dec.Decode(out)
	if err != nil {
		panic(fmt.Errorf(`[gr] failed to XML-decode response body: %w`, err))
	}
	return self
}

/*
Non-panicking version of `(*gr.Res).Xml`. Returns an error if body downloading
or parsing fails. Always closes the body.
*/
func (self *Res) XmlCatch(out interface{}) (err error) {
	defer rec(&err)
	self.Xml(out)
	return
}

/*
Non-panicking version of `(*gr.Res).XmlWith`. Returns an error if body
downloading or parsing fails. Always closes the body.
*/
func (self *Res) XmlWithCatch(out interface{}, fun func(*xml.Decoder)) (err error) {
	defer rec(&err)
	self.XmlWith(out, fun)
	return
}

/*
Returns an error that includes the response HTTP status code and the downloaded
body, as well as the provided short description. Always downloads and closes
the response body, if any. The description must be non-empty, and represent a
reason why the response is unsatisfactory, such as "non-OK" or "non-redirect".
*/
func (self *Res) Err(desc string) Err {
	body := self.Body
	if body == nil {
		return Err{
			Status: self.StatusCode,
			Cause:  errResUnexpectedWithEmptyBody(desc),
		}
	}
	defer body.Close()

	chunk, err := io.ReadAll(body)
	if err != nil {
		return Err{
			Status: self.StatusCode,
			Cause:  errResUnexpectedFailedToReadBody(desc),
		}
	}

	if len(chunk) == 0 {
		return Err{
			Status: self.StatusCode,
			Cause:  errResUnexpectedWithEmptyBody(desc),
		}
	}

	return Err{
		Status: self.StatusCode,
		Body:   chunk,
		Cause:  errResUnexpected(desc),
	}
}

/*
Returns a copy of the response body. Mutates the receiver by fully reading,
closing, and replacing the current body. If the current body is nil, this is a
nop and the output is nil.
*/
func (self *Res) CloneBody() io.ReadCloser {
	one, two := ForkReadCloser(self.Body)
	self.Body = one
	return two
}

/*
Similar to `(*gr.Res).Clone`. Returns a deep copy whose mutations don't affect
the original. Mostly useful for introspection, like dumping to standard output,
which requires reading the body. Clones the body via `(*gr.Res).CloneBody`
which is available separately, and other fields via their standard library
counterparts.
*/
func (self *Res) Clone() *Res {
	if self == nil {
		return nil
	}

	out := *self
	out.Header = self.Header.Clone()
	out.TransferEncoding = cloneStrings(self.TransferEncoding)
	out.Trailer = self.Trailer.Clone()
	out.Request = (*Req)(self.Request).Clone().Req()
	out.Body = self.CloneBody()

	return &out
}

/*
Introspection shortcut. Uses `(*http.Response).Write`, but panics on error
instead of returning an error. Follows the write with a newline. Caution:
mutates the response by reading the body. If you intend to further read the
response, use `.Clone` or `.Dump`.
*/
func (self *Res) Write(out io.Writer) {
	if self != nil && out != nil {
		err := self.Res().Write(out)
		if err != nil {
			panic(errWrite(err))
		}
		_, _ = out.Write(bytesNewline)
	}
}

/*
Introspection tool. Shortcut for using `(*gr.Res).Write` to dump the response to
standard output. Clones before dumping. Can be used in method chains without
affecting the original response.
*/
func (self *Res) Dump() *Res {
	self.Clone().Write(os.Stdout)
	return self
}
