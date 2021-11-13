package gr_test

import (
	"net/http"
	"testing"

	"github.com/mitranim/gr"
)

func TestType(t *testing.T)      { eq(t, `Content-Type`, gr.Type) }
func TestTypeJson(t *testing.T)  { eq(t, `application/x-www-form-urlencoded`, gr.TypeForm) }
func TestTypeForm(t *testing.T)  { eq(t, `application/json`, gr.TypeJson) }
func TestTypeMulti(t *testing.T) { eq(t, `multipart/form-data`, gr.TypeMulti) }

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
