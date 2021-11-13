package gr

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

/*
Alias of `http.Response` with many shortcuts for inspecting and decoding the
response. Freely castable to and from `http.Request`.

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
	panic(self.resErr(`non-OK`))
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
	panic(self.resErr(`non-redirect`))
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
Parses the response body into the given output, which must be either nil or a
pointer. Uses `xml.Decoder` to decode from a stream, without buffering the
entire body. Panics on errors. If the output is nil, skips downloading or
decoding. Returns the same response. Always closes the body.
*/
func (self *Res) Xml(out interface{}) {
	body := self.Body
	if body == nil {
		return
	}
	defer body.Close()

	if out == nil {
		return
	}

	err := xml.NewDecoder(body).Decode(out)
	if err != nil {
		panic(fmt.Errorf(`[gr] failed to XML-decode response body: %w`, err))
	}
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

func (self *Res) resErr(desc string) error {
	body := self.Body
	if body == nil {
		return Err{
			self.StatusCode,
			fmt.Errorf(`unexpected %v response with empty body`, desc),
		}
	}
	defer body.Close()

	chunk, err := io.ReadAll(body)
	if err != nil {
		return Err{
			self.StatusCode,
			fmt.Errorf(`unexpected %v response; failed to read response body: %w`, desc, err),
		}
	}

	if len(chunk) == 0 {
		return Err{
			self.StatusCode,
			fmt.Errorf(`unexpected %v response with empty body`, desc),
		}
	}

	return Err{
		self.StatusCode,
		fmt.Errorf(`unexpected %v response; body: %s`, desc, chunk),
	}
}
