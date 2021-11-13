package gr_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/mitranim/gr"
)

func TestTo(t *testing.T) {
	test := func(val string) {
		t.Helper()
		eq(t, (&gr.Req{URL: &U{Path: val}}), gr.To(val))
	}

	test(``)
	test(`one`)
	test(`/one`)
}

func TestUrl(t *testing.T) {
	test := func(val *U) {
		t.Helper()
		eq(t, (&gr.Req{URL: val}), gr.Url(val))
	}

	test(nil)
	test(&U{Path: `/one`})
}

func TestReq_Ctx_Context(t *testing.T) {
	test := func(val context.Context) {
		t.Helper()

		is(
			t,
			val,
			new(gr.Req).Ctx(val).Context(),
		)
	}

	test(nil)
	test(context.Background())
	test(context.TODO())
}

func TestReq_Meth(t *testing.T) {
	test := func(val string) {
		t.Helper()

		eq(
			t,
			(&gr.Req{Method: val}),
			new(gr.Req).Meth(val),
		)

		eq(
			t,
			(&gr.Req{Method: val}),
			new(gr.Req).Meth(val).Meth(val),
		)

		eq(
			t,
			new(gr.Req),
			new(gr.Req).Meth(val).Meth(``),
		)
	}

	test(``)
	test(http.MethodGet)
	test(http.MethodHead)
	test(http.MethodOptions)
	test(http.MethodPost)
	test(http.MethodPatch)
	test(http.MethodPut)
	test(http.MethodDelete)
}

func TestReq_Get(t *testing.T)     { eq(t, http.MethodGet, new(gr.Req).Get().Method) }
func TestReq_Options(t *testing.T) { eq(t, http.MethodOptions, new(gr.Req).Options().Method) }
func TestReq_Post(t *testing.T)    { eq(t, http.MethodPost, new(gr.Req).Post().Method) }
func TestReq_Patch(t *testing.T)   { eq(t, http.MethodPatch, new(gr.Req).Patch().Method) }
func TestReq_Put(t *testing.T)     { eq(t, http.MethodPut, new(gr.Req).Put().Method) }
func TestReq_Delete(t *testing.T)  { eq(t, http.MethodDelete, new(gr.Req).Delete().Method) }

func TestReq_Path(t *testing.T) {
	test := func(exp U, val string, vals ...interface{}) {
		t.Helper()
		eq(t, (&gr.Req{URL: &exp}), new(gr.Req).Path(val, vals...))
	}

	test(U{}, ``)
	test(U{Path: `/`}, ``, `/`)
	test(U{Path: `/one`}, `/one`)
	test(U{Path: `one/two/0/false`}, `one`, `two`, 0, false)
	test(U{Path: `/one/two/0/false`}, `/one`, `two`, 0, false)
}

// Delegates to `gr.UrlAppend`, so we only need to check the basics.
func TestReq_Append(t *testing.T) {
	test := func(exp string, src *gr.Req, val interface{}) {
		t.Helper()
		eq(t, (&gr.Req{URL: &U{Path: exp}}), src.Append(val))
	}

	test(`/`, new(gr.Req), `/`)
	test(`/`, gr.Path(``), `/`)
	test(`one`, gr.Path(``), `one`)
	test(`one/two`, gr.Path(`one`), `two`)

	panics(t, `[gt] failed to append to URL path: unexpected empty string`, func() {
		new(gr.Req).Append(nil)
	})

	panics(t, `[gt] failed to append to URL path: unexpected empty string`, func() {
		new(gr.Req).Append(``)
	})
}

// Delegates to `gr.UrlJoin`, so we only need to check the basics.
func TestReq_Join(t *testing.T) {
	test := func(exp string, src *gr.Req, vals ...interface{}) {
		t.Helper()
		eq(t, (&gr.Req{URL: &U{Path: exp}}), src.Join(vals...))
	}

	test(``, new(gr.Req))
	test(`one/0/false`, gr.Path(``), `one`, 0, false)
	test(`/one/0/false`, gr.Path(`/`), `one`, 0, false)
	test(`one/two/0/false`, gr.Path(`one`), `two`, 0, false)
	test(`/one/two/0/false`, gr.Path(`/one`), `two`, 0, false)

	panics(t, `[gt] failed to append to URL path: unexpected empty string`, func() {
		new(gr.Req).Join(nil)
	})

	panics(t, `[gt] failed to append to URL path: unexpected empty string`, func() {
		new(gr.Req).Join(``)
	})
}

func TestReq_To(t *testing.T) {
	test := func(exp U, val string) {
		t.Helper()
		eq(t, (&gr.Req{URL: &exp}), new(gr.Req).To(val))
	}

	test(U{}, ``)
	test(U{Path: `/one`}, `/one`)
	test(U{Path: `/one`, RawQuery: `two=three`}, `/one?two=three`)

	panics(
		t,
		`[gr] failed to parse request destination: parse "\n": net/url: invalid control character in URL`,
		func() { new(gr.Req).To("\n") },
	)
}

func TestReq_Url(t *testing.T) {
	test := func(val *U) {
		t.Helper()
		eq(t, (&gr.Req{URL: val}), new(gr.Req).Url(val))
	}

	test(nil)
	test(&U{})
	test(&U{Path: `/one`})
	test(&U{Path: `/one`, RawQuery: `two=three`})
}

func TestReq_RawQuery(t *testing.T) {
	test := func(exp U, val string) {
		t.Helper()
		eq(t, (&gr.Req{URL: &exp}), new(gr.Req).RawQuery(val))
	}

	test(U{}, ``)
	test(U{RawQuery: `?`}, `?`)
	test(U{RawQuery: `two=three`}, `two=three`)
	test(U{RawQuery: `?two=three`}, `?two=three`)
}

func TestReq_Query(t *testing.T) {
	test := func(exp U, val V) {
		t.Helper()
		eq(t, (&gr.Req{URL: &exp}), new(gr.Req).Query(val))
	}

	test(U{}, nil)
	test(U{}, V{})
	test(U{}, V{`one`: nil})
	test(U{}, V{`one`: {}})
	test(U{RawQuery: `one=two`}, V{`one`: {`two`}})

	test(
		U{RawQuery: `one=two&three=four&three=five`},
		V{`one`: {`two`}, `three`: {`four`, `five`}},
	)
}

func TestReq_Head(t *testing.T) {
	test := eqTest(t)

	test(
		(&gr.Req{Header: nil}),
		new(gr.Req).Head(nil),
	)

	test(
		(&gr.Req{Header: nil}),
		(&gr.Req{Header: H{gr.Type: {`one`}}}).Head(nil),
	)

	test(
		(&gr.Req{Header: H{gr.Type: {`one`}}}),
		(&gr.Req{Header: nil}).Head(H{gr.Type: {`one`}}),
	)

	test(
		(&gr.Req{Header: H{gr.Type: {`one`}}}),
		(&gr.Req{Header: H{gr.Type: {`two`}}}).Head(H{gr.Type: {`one`}}),
	)
}

// Delegates to `gr.Head.Add`, so we only need to check the basics.
func TestReq_HeadAdd(t *testing.T) {
	test := func(exp, src H, key, val string) {
		t.Helper()
		eq(t, (&gr.Req{Header: exp}), (&gr.Req{Header: src}).HeadAdd(key, val))
	}

	test(H{`One`: {`two`}}, nil, `one`, `two`)

	test(
		H{`one`: {`two`}, `One`: {`three`}},
		H{`one`: {`two`}},
		`One`,
		`three`,
	)

	test(
		H{`One`: {`two`, `three`, `four`}},
		H{`one`: {`two`}, `One`: {`three`}},
		`one`,
		`four`,
	)
}

// Delegates to `gr.Head.Set`, so we only need to check the basics.
func TestReq_HeadSet(t *testing.T) {
	test := func(exp, src H, key, val string) {
		t.Helper()
		eq(t, (&gr.Req{Header: exp}), (&gr.Req{Header: src}).HeadSet(key, val))
	}

	test(H{`One`: {`two`}}, nil, `one`, `two`)

	test(H{`One`: {`three`}}, H{`one`: {`two`}}, `one`, `three`)

	test(
		H{`One`: {`five`}, `three`: {`four`}},
		H{`one`: {`two`}, `three`: {`four`}},
		`one`,
		`five`,
	)
}

// Delegates to `gr.Head.Replace`, so we only need to check the basics.
func TestReq_HeadReplace(t *testing.T) {
	test := func(exp, src H, key string, vals ...string) {
		t.Helper()
		eq(t, (&gr.Req{Header: exp}), (&gr.Req{Header: src}).HeadReplace(key, vals...))
	}

	test(nil, nil, `one`)

	test(H{}, H{}, `one`)

	test(
		H{`one`: {`two`}},
		H{`one`: {`two`}},
		`three`,
	)

	test(
		H{`One`: {`two`}},
		nil,
		`one`, `two`,
	)

	test(
		H{`One`: {`two`}},
		nil,
		`One`, `two`,
	)

	test(
		H{`One`: {`two`}},
		H{},
		`one`, `two`,
	)

	test(
		H{},
		H{`one`: nil},
		`one`,
	)

	test(
		H{},
		H{`One`: nil},
		`one`,
	)

	test(
		H{},
		H{`one`: {}},
		`one`,
	)

	test(
		H{},
		H{`One`: {}},
		`one`,
	)

	test(
		H{},
		H{`one`: {`two`}},
		`one`,
	)

	test(
		H{},
		H{`One`: {`two`}},
		`one`,
	)

	test(
		H{`Two`: {`three`}},
		H{`One`: {`two`}, `Two`: {`three`}},
		`one`,
	)

	test(
		H{`Two`: {`three`}},
		H{`One`: {`two`}, `Two`: {`three`}},
		`One`,
	)

	test(
		H{`One`: {`two`}},
		H{`one`: {`three`}},
		`one`, `two`,
	)

	test(
		H{`One`: {`two`, `three`}},
		nil,
		`one`, `two`, `three`,
	)

	test(
		H{`One`: {`two`, `three`}},
		H{},
		`one`, `two`, `three`,
	)

	test(
		H{`one`: {`two`}, `Three`: {`five`}},
		H{`one`: {`two`}, `three`: {`four`}},
		`three`, `five`,
	)
}

// Delegates to `gr.Head.Patch`, so we only need to check the basics.
func TestReq_HeadPatch(t *testing.T) {
	test := func(exp, src, patch H) {
		t.Helper()
		eq(t, (&gr.Req{Header: exp}), (&gr.Req{Header: src}).HeadPatch(patch))
	}

	test(nil, nil, nil)
	test(nil, nil, H{})
	test(H{}, H{}, nil)
	test(H{}, H{}, H{})

	test(
		H{`One`: {`two`}},
		nil,
		H{`one`: {`two`}},
	)

	test(
		H{`One`: {`five`}, `three`: {`four`}, `Six`: {`seven`}},
		H{`one`: {`two`}, `three`: {`four`}},
		H{`one`: {`five`}, `six`: {`seven`}},
	)
}

func TestReq_Type(t *testing.T) {
	test := func(exp, src H, typ string) {
		t.Helper()
		eq(t, (&gr.Req{Header: exp}), (&gr.Req{Header: src}).Type(typ))
	}

	test(nil, nil, ``)
	test(H{}, H{}, ``)

	test(
		H{`content-type`: {`one`}},
		H{`content-type`: {`one`}},
		``,
	)

	test(
		H{`content-type`: {`one`}, gr.Type: {`two`}},
		H{`content-type`: {`one`}},
		`two`,
	)

	test(
		H{gr.Type: {`two`}},
		H{gr.Type: {`one`}},
		`two`,
	)

	test(H{}, H{gr.Type: {`one`}}, ``)
}

func TestReq_String(t *testing.T) {
	eq(t, new(gr.Req), new(gr.Req).String(``))

	testBodyString(
		t,
		`hello world`,
		new(gr.Req).String(`hello world`),
	)
}

func testBodyString(t testing.TB, src string, req *gr.Req) {
	eq(t, int64(len(src)), req.ContentLength)
	eq(t, gr.NewStringReadCloser(src), req.Body)
	testGetBody(t, src, gr.NewStringReadCloser(src), req)
	testGetBody(t, src, gr.NewStringReadCloser(src), req)
}

func testGetBody(t testing.TB, text string, body io.ReadCloser, req *gr.Req) {
	reader, err := req.GetBody()
	eq(t, nil, err)

	chunk, err := io.ReadAll(reader)
	eq(t, nil, err)
	eq(t, text, string(chunk))
}

func TestReq_Bytes(t *testing.T) {
	eq(t, new(gr.Req), new(gr.Req).Bytes(nil))
	eq(t, new(gr.Req), new(gr.Req).Bytes([]byte{}))

	testBodyBytes(
		t,
		`hello world`,
		new(gr.Req).Bytes([]byte(`hello world`)),
	)
}

func testBodyBytes(t testing.TB, src string, req *gr.Req) {
	eq(t, int64(len(src)), req.ContentLength)
	eq(t, gr.NewBytesReadCloser([]byte(src)), req.Body)
	testGetBody(t, src, gr.NewBytesReadCloser([]byte(src)), req)
	testGetBody(t, src, gr.NewBytesReadCloser([]byte(src)), req)
}

func TestReq_Vals(t *testing.T) {
	eq(t, new(gr.Req), new(gr.Req).Vals(nil))
	eq(t, new(gr.Req), new(gr.Req).Vals(V{}))
	testBodyString(t, `one=two`, new(gr.Req).Vals(V{`one`: {`two`}}))
}

func TestReq_FormVals(t *testing.T) {
	test := eqTest(t)

	test(
		(&gr.Req{Header: H{gr.Type: {gr.TypeForm}}}),
		new(gr.Req).FormVals(nil),
	)

	test(
		(&gr.Req{Header: H{gr.Type: {gr.TypeForm}}}),
		new(gr.Req).FormVals(V{}),
	)

	test(
		(&gr.Req{Method: http.MethodGet, Header: H{gr.Type: {gr.TypeForm}}}),
		new(gr.Req).Get().FormVals(nil),
	)

	test(
		(&gr.Req{Method: http.MethodGet, Header: H{gr.Type: {gr.TypeForm}}}),
		new(gr.Req).Get().FormVals(V{}),
	)

	test(
		(&gr.Req{Method: http.MethodPost, Header: H{gr.Type: {gr.TypeForm}}}),
		new(gr.Req).Post().FormVals(nil),
	)

	test(
		(&gr.Req{Method: http.MethodPost, Header: H{gr.Type: {gr.TypeForm}}}),
		new(gr.Req).Post().FormVals(V{}),
	)

	t.Run(`non-empty`, func(t *testing.T) {
		req := new(gr.Req).FormVals(V{`one`: {`two`}})
		testBodyString(t, `one=two`, req)
		eq(t, H{gr.Type: {gr.TypeForm}}, req.Header)
	})
}

func TestReq_Json(t *testing.T) {
	t.Run(`empty`, func(t *testing.T) {
		eq(
			t,
			(&gr.Req{Header: H{gr.Type: {gr.TypeJson}}}),
			new(gr.Req).Json(nil),
		)
	})

	t.Run(`GET empty`, func(t *testing.T) {
		eq(
			t,
			(&gr.Req{Method: http.MethodGet, Header: H{gr.Type: {gr.TypeJson}}}),
			new(gr.Req).Get().Json(nil),
		)
	})

	t.Run(`POST empty`, func(t *testing.T) {
		req := new(gr.Req).Post().Json(nil)
		testBodyBytes(t, `null`, req)
		eq(t, http.MethodPost, req.Method)
		eq(t, H{gr.Type: {gr.TypeJson}}, req.Header)
	})

	t.Run(`GET non-empty`, func(t *testing.T) {
		req := new(gr.Req).Get().Json(json.RawMessage(`"hello world"`))
		testBodyBytes(t, `"hello world"`, req)
		eq(t, http.MethodGet, req.Method)
		eq(t, H{gr.Type: {gr.TypeJson}}, req.Header)
	})

	t.Run(`POST non-empty`, func(t *testing.T) {
		req := new(gr.Req).Post().Json(json.RawMessage(`"hello world"`))
		testBodyBytes(t, `"hello world"`, req)
		eq(t, http.MethodPost, req.Method)
		eq(t, H{gr.Type: {gr.TypeJson}}, req.Header)
	})
}

func TestReq_ReadCloser(t *testing.T) {
	eq(t, new(gr.Req), new(gr.Req).ReadCloser(nil))

	eq(
		t,
		(&gr.Req{Body: gr.NewStringReadCloser(`hello world`)}),
		new(gr.Req).ReadCloser(gr.NewStringReadCloser(`hello world`)),
	)
}

func TestReq_Reader(t *testing.T) {
	eq(t, new(gr.Req), new(gr.Req).Reader(nil))

	eq(
		t,
		(&gr.Req{Body: io.NopCloser(strings.NewReader(`hello world`))}),
		new(gr.Req).Reader(strings.NewReader(`hello world`)),
	)
}

func TestReq_Clone(t *testing.T) {
	srcUrl := &U{Path: `/one`}
	srcHead := H{gr.Type: {gr.TypeForm}}
	src := new(gr.Req).Url(srcUrl).Head(srcHead)

	tar := src.Clone(context.Background())
	tar.URL.Path = `/two`
	tar.Header.Set(gr.Type, gr.TypeJson)

	eq(
		t,
		(&gr.Req{
			URL:    &U{Path: `/one`},
			Header: H{gr.Type: {gr.TypeForm}},
		}),
		src,
	)

	eq(
		t,
		(&gr.Req{
			URL:    &U{Path: `/two`},
			Header: H{gr.Type: {gr.TypeJson}},
		}).Ctx(context.Background()),
		tar,
	)
}

func TestReq_Init(t *testing.T) {
	test := eqTest(t)

	test(
		(&gr.Req{
			Method: http.MethodGet,
			Header: H{},
			URL:    &U{},
		}).Ctx(context.Background()),
		new(gr.Req).Init(),
	)

	test(
		(&gr.Req{
			Method: http.MethodPost,
			Header: H{},
			URL:    &U{},
		}).Ctx(context.Background()),
		new(gr.Req).Post().Init(),
	)

	test(
		(&gr.Req{
			Method: http.MethodGet,
			Header: H{gr.Type: {gr.TypeForm}},
			URL:    &U{},
		}).Ctx(context.Background()),
		new(gr.Req).TypeForm().Init(),
	)

	test(
		(&gr.Req{
			Method: http.MethodPost,
			Header: H{gr.Type: {gr.TypeForm}},
			URL:    &U{},
		}).Ctx(context.Background()),
		new(gr.Req).Post().TypeForm().Init(),
	)

	test(
		(&gr.Req{
			Method: http.MethodGet,
			Header: H{},
			URL:    &U{Path: `/one`},
		}).Ctx(context.Background()),
		gr.To(`/one`).Init(),
	)

	test(
		(&gr.Req{
			Method: http.MethodPost,
			Header: H{},
			URL:    &U{Path: `/one`},
		}).Ctx(context.Background()),
		gr.To(`/one`).Post().Init(),
	)

	test(
		(&gr.Req{
			Method: http.MethodPost,
			Header: H{gr.Type: {gr.TypeForm}},
			URL:    &U{Path: `/one`},
		}).Ctx(context.Background()),
		gr.To(`/one`).Post().TypeForm().Init(),
	)
}

func TestReq_Req(t *testing.T) {
	eq(t, new(Q), new(gr.Req).Req())
}

func TestReq_CliRes(t *testing.T) {
	trans := Trans{
		Res: &S{Body: gr.NewStringReadCloser(`hello world`)},
	}

	res := new(gr.Req).CliRes(&http.Client{Transport: &trans})

	eq(t, gr.Init().Req(), trans.Req)
	eq(t, trans.Res, res.Res())
}
