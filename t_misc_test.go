package gr_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/mitranim/gr"
)

func TestType(t *testing.T) {
	eq(t, `Content-Type`, gr.Type)
}

func TestTypeJson(t *testing.T) {
	eq(t, `application/x-www-form-urlencoded; charset=utf-8`, gr.TypeForm)
}

func TestTypeForm(t *testing.T) {
	eq(t, `application/json; charset=utf-8`, gr.TypeJson)
}

func TestTypeMulti(t *testing.T) {
	eq(t, `multipart/form-data; charset=utf-8`, gr.TypeMulti)
}

func TestIsReadOnly(t *testing.T) {
	test := func(exp bool, val string) {
		t.Helper()
		eq(t, exp, gr.IsReadOnly(val))
	}

	test(true, ``)
	test(true, `GET`)
	test(true, `HEAD`)
	test(true, `OPTIONS`)
	test(false, `POST`)
	test(false, `PUT`)
	test(false, `PATCH`)
	test(false, `DELETE`)
	test(false, `CONNECT`)
	test(false, `TRACE`)

	test(true, http.MethodGet)
	test(true, http.MethodHead)
	test(true, http.MethodOptions)
	test(false, http.MethodPost)
	test(false, http.MethodPut)
	test(false, http.MethodPatch)
	test(false, http.MethodDelete)
	test(false, http.MethodConnect)
	test(false, http.MethodTrace)
}

func TestIsInfo(t *testing.T)      { testIsInfo(t, gr.IsInfo) }
func TestIsOk(t *testing.T)        { testIsOk(t, gr.IsOk) }
func TestIsRedir(t *testing.T)     { testIsRedir(t, gr.IsRedir) }
func TestIsClientErr(t *testing.T) { testIsClientErr(t, gr.IsClientErr) }
func TestIsServerErr(t *testing.T) { testIsServerErr(t, gr.IsServerErr) }

func testIsInfo(t testing.TB, fun func(int) bool) {
	test := codeTest(t, fun)

	test(false, 0)
	test(false, 1)
	test(false, 99)
	test(true, 100)
	test(true, 101)
	test(true, 199)
	test(false, 200)
	test(false, 201)
	test(false, 299)
	test(false, 300)
	test(false, 301)
	test(false, 399)
	test(false, 400)
	test(false, 401)
	test(false, 499)
	test(false, 500)
	test(false, 501)
	test(false, 599)
	test(false, 600)
}

func testIsOk(t testing.TB, fun func(int) bool) {
	test := codeTest(t, fun)

	test(false, 0)
	test(false, 1)
	test(false, 99)
	test(false, 100)
	test(false, 101)
	test(false, 199)
	test(true, 200)
	test(true, 201)
	test(true, 299)
	test(false, 300)
	test(false, 301)
	test(false, 399)
	test(false, 400)
	test(false, 401)
	test(false, 499)
	test(false, 500)
	test(false, 501)
	test(false, 599)
	test(false, 600)
}

func testIsRedir(t testing.TB, fun func(int) bool) {
	test := codeTest(t, fun)

	test(false, 0)
	test(false, 1)
	test(false, 99)
	test(false, 100)
	test(false, 101)
	test(false, 199)
	test(false, 200)
	test(false, 201)
	test(false, 299)
	test(true, 300)
	test(true, 301)
	test(true, 399)
	test(false, 400)
	test(false, 401)
	test(false, 499)
	test(false, 500)
	test(false, 501)
	test(false, 599)
	test(false, 600)
}

func testIsClientErr(t testing.TB, fun func(int) bool) {
	test := codeTest(t, fun)

	test(false, 0)
	test(false, 1)
	test(false, 99)
	test(false, 100)
	test(false, 101)
	test(false, 199)
	test(false, 200)
	test(false, 201)
	test(false, 299)
	test(false, 300)
	test(false, 301)
	test(false, 399)
	test(true, 400)
	test(true, 401)
	test(true, 499)
	test(false, 500)
	test(false, 501)
	test(false, 599)
	test(false, 600)
}

func testIsServerErr(t testing.TB, fun func(int) bool) {
	test := codeTest(t, fun)

	test(false, 0)
	test(false, 1)
	test(false, 99)
	test(false, 100)
	test(false, 101)
	test(false, 199)
	test(false, 200)
	test(false, 201)
	test(false, 299)
	test(false, 300)
	test(false, 301)
	test(false, 399)
	test(false, 400)
	test(false, 401)
	test(false, 499)
	test(true, 500)
	test(true, 501)
	test(true, 599)
	test(false, 600)
}

func TestStr(t *testing.T) {
	test := func(exp string, src interface{}) {
		t.Helper()
		eq(t, exp, gr.Str(src))
	}

	test(``, nil)
	test(``, ``)
	test(``, (*string)(nil))
	test(``, (**string)(nil))
	test(``, (*int)(nil))
	test(``, (**int)(nil))
	test(`one`, `one`)
	test(`0`, 0)
	test(`10`, 10)
	test(`123.456`, 123.456)
	test(`true`, true)
	test(`false`, false)
	test(`0001-02-03 04:05:06.000000007 +0000 UTC`, time.Date(1, 2, 3, 4, 5, 6, 7, time.UTC))
	test(``, &url.URL{})
	test(`/one`, &url.URL{Path: `/one`})

	panics(t, `unsupported type "[]int" of kind "slice"`, func() {
		gr.Str([]int(nil))
	})

	panics(t, `unsupported type "struct {}" of kind "struct"`, func() {
		gr.Str(struct{}{})
	})

	panics(t, `unsupported type "map[int]int" of kind "map"`, func() {
		gr.Str((map[int]int)(nil))
	})

	panics(t, `unsupported type "func(interface {}) string" of kind "func"`, func() {
		gr.Str(gr.Str)
	})

	panics(t, `unsupported type "chan int" of kind "chan"`, func() {
		gr.Str((chan int)(nil))
	})
}

func TestUrlAppend(t *testing.T) {
	test := func(exp, src *url.URL, val interface{}) {
		t.Helper()

		out := gr.UrlAppend(src, val)
		eq(t, exp, out)

		if src != nil {
			is(t, src, out)
		}
	}

	testOne := func(exp string, val interface{}) {
		t.Helper()
		test(&url.URL{Path: exp}, nil, val)
	}

	panics(
		t,
		`[gt] failed to append to URL path: unexpected empty string`,
		func() { testOne(``, ``) },
	)

	panics(
		t,
		`[gt] failed to append to URL path: unexpected empty string`,
		func() { testOne(``, &url.URL{}) },
	)

	testOne(`10`, 10)
	testOne(`true`, true)
	testOne(`/`, `/`)
	testOne(`one`, `one`)
	testOne(`/one`, `/one`)
	testOne(`one/`, `one/`)
	testOne(`one/two`, `one/two`)
	testOne(`/one/two`, `/one/two`)
	testOne(`/one/two/`, `/one/two/`)
	testOne(`one/two/`, `one/two/`)

	testTwo := func(exp, src string, val interface{}) {
		t.Helper()

		test(
			&url.URL{Path: exp},
			&url.URL{Path: src},
			val,
		)

		const query = `3d0abf=2a6bab`
		const frag = `46343b`

		test(
			&url.URL{Path: exp, RawQuery: query, Fragment: frag},
			&url.URL{Path: src, RawQuery: query, Fragment: frag},
			val,
		)
	}

	panics(
		t,
		`[gt] failed to append to URL path: unexpected empty string`,
		func() { testTwo(``, ``, ``) },
	)

	panics(
		t,
		`[gt] failed to append to URL path: unexpected empty string`,
		func() { testTwo(``, ``, &url.URL{}) },
	)

	testTwo(`10`, ``, 10)
	testTwo(`true`, ``, true)
	testTwo(`/`, ``, `/`)
	testTwo(`/`, `/`, `/`)
	testTwo(`/`, `//`, `/`)
	testTwo(`/`, `/`, `//`)
	testTwo(`/`, `//`, `//`)
	testTwo(`one`, `one`, `/`)
	testTwo(`one/10`, `one`, 10)
	testTwo(`one/true`, `one`, true)
	testTwo(`one/two`, `one`, &url.URL{Path: `two`})
	testTwo(`one/two`, `one`, &url.URL{Path: `/two`})
	testTwo(`one`, `one/`, `/`)
	testTwo(`one/two`, `one`, `two`)
	testTwo(`one/two`, `one`, `/two`)
	testTwo(`one/two`, `one/`, `two`)
	testTwo(`one/two`, `one/`, `/two`)
	testTwo(`one/two`, `one/`, `/two/`)
	testTwo(`one/two/three`, `one/`, `/two/three`)
	testTwo(`one/two/three`, `one/`, `/two/three/`)
	testTwo(`one/two/three`, `one`, `/two/three/`)
	testTwo(`one/two/three`, `one`, `two/three/`)
	testTwo(`one/two/three`, `one/`, `two/three/`)
}

// Defers to `UrlAppend`, so we only need to check the basics.
func TestUrlJoin(t *testing.T) {
	testOk := func(exp, init string, vals ...interface{}) {
		t.Helper()

		src := &url.URL{Path: init}
		out := gr.UrlJoin(src, vals...)

		eq(t, &url.URL{Path: exp}, out)

		if src != nil {
			is(t, src, out)
		}
	}

	testOk(``, ``)
	testOk(`0`, ``, 0)
	testOk(`one/0`, `one`, 0)
	testOk(`one/0/false`, `one`, 0, false)
	testOk(`10/true/one`, ``, 10, `true`, `one`)
	testOk(`one/10/true/two`, `one`, 10, `true`, `two`)

	testFail := func(init string, vals ...interface{}) {
		t.Helper()

		panics(
			t,
			`[gt] failed to append to URL path: unexpected empty string`,
			func() { gr.UrlJoin(&url.URL{Path: init}, vals...) },
		)
	}

	testFail(``, nil)
	testFail(``, ``)
	testFail(``, `one`, nil)
	testFail(``, `one`, nil, `two`)
	testFail(``, `one`, ``, `two`)
	testFail(`one`, nil)
	testFail(`one`, ``)
	testFail(`one`, `two`, nil)
	testFail(`one`, `two`, nil, `three`)
	testFail(`one`, `two`, ``, `three`)
	testFail(`one`, `two`, &url.URL{}, `three`)
}
