## Overview

"gr" stands for for **G**o **R**equest or **Go** **R**equest-**R**esponse. It also represents my reaction to many APIs in "net/http". Shortcuts for making HTTP requests and reading HTTP responses. Features:

  * Brevity!
  * **No added wrappers or interfaces**. Just aliases for `http.Request` and `http.Response`, freely-castable back and forth.
  * Fluent chainable builder API.
  * Most methods are nullary or unary. No syntactic overhead for features you don't use.
  * No overhead over "lower-level" uses of `http.Request` and `http.Response`. Many shortcuts are more performant than "standard" approaches.
  * Resulting requests can be passed to any code that takes `http.Request`.
  * Usable for `http.Response` responses obtained from any source.
  * Tiny and dependency-free.

API docs: https://pkg.go.dev/github.com/mitranim/gr.

## Usage

Sending and receiving JSON:

```golang
package gr_test

import (
  "fmt"
  "net/url"

  "github.com/mitranim/gr"
)

func ExampleReq_jsonInputJsonOutput() {
  input := Input{`some input`}

  var output Output
  gr.To(testServer.URL).Path(`/json`).Json(input).Res().Ok().Json(&output)

  fmt.Printf("%#v\n", output)

  // Output:
  // gr_test.Output{ReqMethod:"GET", ReqUrl:"/json", ReqBody:"{\"inputVal\":\"some input\"}"}
}

type Input struct {
  InputVal string `json:"inputVal"`
}

type Output struct {
  ReqMethod string `json:"reqMethod"`
  ReqUrl    string `json:"reqUrl"`
  ReqBody   string `json:"reqBody"`
}
```

Sending URL-encoded form, reading plain text:

```golang
package gr_test

import (
  "fmt"
  "net/url"

  "github.com/mitranim/gr"
)

func ExampleReq_formBodyPlainResponse() {
  req := gr.To(testServer.URL).Post().FormVals(url.Values{`one`: {`two`}})
  res := req.Res().Ok()
  defer res.Done()

  fmt.Printf("\nresponse status: %v\n", res.StatusCode)
  fmt.Printf("\nresponse type: %v\n", res.Header.Get(gr.Type))
  fmt.Printf("\nresponse body:\n%v\n", res.ReadString())

  // Output:
  //
  // response status: 200
  //
  // response type: text/plain; charset=utf-8
  //
  // response body:
  //
  // request method: POST
  // request URL: /
  // request body: one=two
}
```

## Why pointers

Since `gr` uses a chainable builder-style API, it could have defined all "builder" methods on `Req` and `Res`, rather than `*Req` and `*Res`. This would allow to store "partially built" requests, and "fork" them by simply reassigning the variable. So why pointers instead?

* In Go, request and response are inherently mutable, because they contain reference types such as `*url.URL`, `http.Header` and `io.ReadCloser`. Copying the struct `http.Request` or `http.Response` by reassigning the variable makes a shallow copy, while the inner references are still shared. That would be hazardous. Only explicit copying via `.Clone` is viable.

* Emulating an "immutable" API by using copy-on-write for URL and headers is possible, but incurs a measurable performance penalty.

* All APIs in `"net/http"` operate on requests and responses by pointer. By using the same pointers, we avoid the overhead of copying and reallocation.

* Go request and response structs are rather large. The language seems to use naive call conventions that involve always copying value types, as opposed to passing them by reference when they would be "const". For large structs, always passing them by pointer rather than by value seems faster.

## License

https://unlicense.org

## Misc

I'm receptive to suggestions. If this library _almost_ satisfies you but needs changes, open an issue or chat me up. Contacts: https://mitranim.com/#contacts
