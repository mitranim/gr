package gr_test

import (
	"net/url"
	"testing"

	"github.com/mitranim/gr"
)

func TestRes_IsInfo(t *testing.T) {
	testIsInfo(t, func(code int) bool {
		return (&gr.Res{StatusCode: code}).IsInfo()
	})
}

func TestRes_IsOk(t *testing.T) {
	testIsOk(t, func(code int) bool {
		return (&gr.Res{StatusCode: code}).IsOk()
	})
}

func TestRes_IsRedir(t *testing.T) {
	testIsRedir(t, func(code int) bool {
		return (&gr.Res{StatusCode: code}).IsRedir()
	})
}

func TestRes_IsClientErr(t *testing.T) {
	testIsClientErr(t, func(code int) bool {
		return (&gr.Res{StatusCode: code}).IsClientErr()
	})
}

func TestRes_IsServerErr(t *testing.T) {
	testIsServerErr(t, func(code int) bool {
		return (&gr.Res{StatusCode: code}).IsServerErr()
	})
}

func TestRes_OkCatch(t *testing.T) {
	eq(t, nil, (&gr.Res{StatusCode: 200}).OkCatch())
	eq(t, nil, (&gr.Res{StatusCode: 201}).OkCatch())
	eq(t, nil, (&gr.Res{StatusCode: 299}).OkCatch())
	eq(t, nil, (&gr.Res{StatusCode: 200, Body: gr.NewStringReadCloser(`hello world`)}).OkCatch())
	eq(t, nil, (&gr.Res{StatusCode: 201, Body: gr.NewStringReadCloser(`hello world`)}).OkCatch())
	eq(t, nil, (&gr.Res{StatusCode: 299, Body: gr.NewStringReadCloser(`hello world`)}).OkCatch())

	errs(
		t,
		`[gr] error: unexpected non-OK response with empty body`,
		new(gr.Res).OkCatch(),
	)

	errs(
		t,
		`[gr] error (HTTP status 101): unexpected non-OK response with empty body`,
		(&gr.Res{StatusCode: 101}).OkCatch(),
	)

	errs(
		t,
		`[gr] error: unexpected non-OK response; body: hello world`,
		(&gr.Res{Body: gr.NewStringReadCloser(`hello world`)}).OkCatch(),
	)

	errs(
		t,
		`[gr] error (HTTP status 101): unexpected non-OK response; body: hello world`,
		(&gr.Res{StatusCode: 101, Body: gr.NewStringReadCloser(`hello world`)}).OkCatch(),
	)

	t.Run(`closing`, func(t *testing.T) {
		t.Run(`success`, func(t *testing.T) {
			eq(
				t,
				nil,
				(&gr.Res{StatusCode: 200, Body: FailReadCloser{}}).OkCatch(),
			)
		})

		t.Run(`empty body`, func(t *testing.T) {
			body := new(ReadCloseFlag)

			errs(
				t,
				`[gr] error (HTTP status 101): unexpected non-OK response with empty body`,
				(&gr.Res{StatusCode: 101, Body: body}).OkCatch(),
			)

			eq(
				t,
				&ReadCloseFlag{DidRead: true, DidClose: true},
				body,
			)
		})

		t.Run(`non-empty body`, func(t *testing.T) {
			body := NewReaderCloseFlag(`hello world`)

			errs(
				t,
				`[gr] error (HTTP status 101): unexpected non-OK response; body: hello world`,
				(&gr.Res{StatusCode: 101, Body: body}).OkCatch(),
			)

			eq(t, true, body.DidClose)
		})
	})
}

func TestRes_RedirCatch(t *testing.T) {
	eq(t, nil, (&gr.Res{StatusCode: 300}).RedirCatch())
	eq(t, nil, (&gr.Res{StatusCode: 301}).RedirCatch())
	eq(t, nil, (&gr.Res{StatusCode: 399}).RedirCatch())

	errs(
		t,
		`[gr] error: unexpected non-redirect response with empty body`,
		new(gr.Res).RedirCatch(),
	)

	errs(
		t,
		`[gr] error (HTTP status 101): unexpected non-redirect response with empty body`,
		(&gr.Res{StatusCode: 101}).RedirCatch(),
	)

	errs(
		t,
		`[gr] error: unexpected non-redirect response; body: hello world`,
		(&gr.Res{Body: gr.NewStringReadCloser(`hello world`)}).RedirCatch(),
	)

	errs(
		t,
		`[gr] error (HTTP status 101): unexpected non-redirect response; body: hello world`,
		(&gr.Res{StatusCode: 101, Body: gr.NewStringReadCloser(`hello world`)}).RedirCatch(),
	)

	t.Run(`closing`, func(t *testing.T) {
		t.Run(`success`, func(t *testing.T) {
			eq(
				t,
				nil,
				(&gr.Res{StatusCode: 300, Body: FailReadCloser{}}).RedirCatch(),
			)
		})

		t.Run(`empty body`, func(t *testing.T) {
			body := new(ReadCloseFlag)

			errs(
				t,
				`[gr] error (HTTP status 101): unexpected non-redirect response with empty body`,
				(&gr.Res{StatusCode: 101, Body: body}).RedirCatch(),
			)

			eq(
				t,
				&ReadCloseFlag{DidRead: true, DidClose: true},
				body,
			)
		})

		t.Run(`non-empty body`, func(t *testing.T) {
			body := NewReaderCloseFlag(`hello world`)

			errs(
				t,
				`[gr] error (HTTP status 101): unexpected non-redirect response; body: hello world`,
				(&gr.Res{StatusCode: 101, Body: body}).RedirCatch(),
			)

			eq(t, true, body.DidClose)
		})
	})
}

func TestRes_CloseErr(t *testing.T) {
	eq(t, nil, new(gr.Res).CloseErr())

	body := &ReadCloseFlag{}

	eq(t, nil, (&gr.Res{Body: body}).CloseErr())
	eq(t, &ReadCloseFlag{DidRead: false, DidClose: true}, body)
}

func TestRes_ReadBytes(t *testing.T) {
	eq(t, []byte(nil), new(gr.Res).ReadBytes())

	body := NewReaderCloseFlag(`hello world`)

	eq(t, `hello world`, string((&gr.Res{Body: body}).ReadBytes()))
	eq(t, true, body.DidClose)

	eq(t, []byte{}, (&gr.Res{Body: ReaderFailCloser{}}).ReadBytes())
}

func TestRes_ReadString(t *testing.T) {
	eq(t, ``, new(gr.Res).ReadString())

	body := NewReaderCloseFlag(`hello world`)

	eq(t, `hello world`, string((&gr.Res{Body: body}).ReadString()))
	eq(t, true, body.DidClose)

	eq(t, ``, (&gr.Res{Body: ReaderFailCloser{}}).ReadString())
}

func TestRes_Form(t *testing.T) {
	eq(t, url.Values(nil), new(gr.Res).Form())

	t.Run(`decode and close body`, func(t *testing.T) {
		body := NewReaderCloseFlag(`one=two&three=four&three=five`)

		eq(
			t,
			url.Values{`one`: {`two`}, `three`: {`four`, `five`}},
			(&gr.Res{Body: body}).Form(),
		)
		eq(t, true, body.DidClose)
	})

	t.Run(`catch decoding error and close body`, func(t *testing.T) {
		body := NewReaderCloseFlag(`;`)
		val, err := (&gr.Res{Body: body}).FormCatch()

		eq(t, url.Values(nil), val)
		errs(t, `[gr] failed to form-decode response body: invalid semicolon separator in query`, err)
		eq(t, true, body.DidClose)
	})
}

func TestRes_JsonCatch(t *testing.T) {
	eq(t, nil, new(gr.Res).JsonCatch(nil))

	t.Run(`allow nil output and close body`, func(t *testing.T) {
		body := &ReadCloseFlag{}
		eq(t, nil, (&gr.Res{Body: body}).JsonCatch(nil))
		eq(t, &ReadCloseFlag{DidRead: false, DidClose: true}, body)
	})

	t.Run(`decode and close body`, func(t *testing.T) {
		body := NewReaderCloseFlag(`"hello world"`)

		var out string
		eq(t, nil, (&gr.Res{Body: body}).JsonCatch(&out))

		eq(t, `hello world`, out)
		eq(t, true, body.DidClose)
	})

	t.Run(`catch decoding error and close body`, func(t *testing.T) {
		body := NewReaderCloseFlag(`null`)

		errs(
			t,
			`[gr] failed to JSON-decode response body: failing to decode`,
			(&gr.Res{Body: body}).JsonCatch(new(DecodeFail)),
		)

		eq(t, true, body.DidClose)
	})
}

func TestRes_JsonEither(t *testing.T) {
	eq(t, false, new(gr.Res).JsonEither(nil, nil))
	eq(t, false, (&gr.Res{StatusCode: 199}).JsonEither(nil, nil))
	eq(t, true, (&gr.Res{StatusCode: 200}).JsonEither(nil, nil))
	eq(t, false, (&gr.Res{StatusCode: 300}).JsonEither(nil, nil))

	testJsonEither(t, Pair{nil, nil}, 0, Pair{nil, nil})
	testJsonEither(t, Pair{nil, nil}, 200, Pair{nil, nil})
	testJsonEither(t, Pair{nil, spr(`hello world`)}, 0, Pair{nil, spr(``)})
	testJsonEither(t, Pair{spr(`hello world`), nil}, 200, Pair{spr(``), nil})
	testJsonEither(t, Pair{spr(``), spr(`hello world`)}, 0, Pair{spr(``), spr(``)})
	testJsonEither(t, Pair{spr(`hello world`), spr(``)}, 200, Pair{spr(``), spr(``)})
}

func testJsonEither(t testing.TB, exp Pair, status int, inp Pair) {
	t.Helper()

	body := NewReaderCloseFlag(`"hello world"`)
	(&gr.Res{StatusCode: status, Body: body}).JsonEither(inp[0], inp[1])

	eq(t, exp[0], inp[0])
	eq(t, exp[1], inp[1])
	eq(t, true, body.DidClose)
}

func TestRes_XmlCatch(t *testing.T) {
	eq(t, nil, new(gr.Res).XmlCatch(nil))

	t.Run(`allow nil output and close body`, func(t *testing.T) {
		body := &ReadCloseFlag{}
		eq(t, nil, (&gr.Res{Body: body}).XmlCatch(nil))
		eq(t, &ReadCloseFlag{DidRead: false, DidClose: true}, body)
	})

	t.Run(`decode and close body`, func(t *testing.T) {
		body := NewReaderCloseFlag(
			`<string>hello world</string>`,
		)

		var out string
		eq(t, nil, (&gr.Res{Body: body}).XmlCatch(&out))

		eq(t, `hello world`, out)
		eq(t, true, body.DidClose)
	})

	t.Run(`catch decoding error and close body`, func(t *testing.T) {
		body := NewReaderCloseFlag(
			`<string>hello world</string>`,
		)

		errs(
			t,
			`[gr] failed to XML-decode response body: failing to decode`,
			(&gr.Res{Body: body}).XmlCatch(new(DecodeFail)),
		)

		eq(t, true, body.DidClose)
	})
}
