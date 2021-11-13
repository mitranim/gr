package gr_test

import (
	"testing"

	"github.com/mitranim/gr"
)

func TestHead_Init(t *testing.T) {
	type H = gr.Head

	eq(t, H{}, gr.Head(nil).Init())

	head := H{}
	is(t, head, head.Init())
}

func TestHead_Clone(t *testing.T) {
	type H = gr.Head

	src := H{`one`: {`two`}}
	tar := src.Clone()

	tar[`one`][0] = `three`
	tar[`four`] = []string{`five`}

	eq(t, H{`one`: {`two`}}, src)
	eq(t, H{`one`: {`three`}, `four`: {`five`}}, tar)
}

func TestHead_Get(t *testing.T) {
	type H = gr.Head

	eq(t, ``, gr.Head(nil).Get(`one`))
	eq(t, ``, H{`one`: nil, `One`: nil}.Get(`one`))
	eq(t, ``, H{`one`: {}, `One`: {}}.Get(`one`))
	eq(t, `two`, H{`one`: {`two`}, `One`: {`three`}}.Get(`one`))
	eq(t, `three`, H{`One`: {`three`}}.Get(`one`))
	eq(t, `three`, H{`one`: {`two`}, `One`: {`three`}}.Get(`One`))
}

func TestHead_Values(t *testing.T) {
	type H = gr.Head

	eq(t, []string(nil), gr.Head(nil).Values(`one`))
	eq(t, []string(nil), H{`one`: nil, `One`: nil}.Values(`one`))
	eq(t, []string{}, H{`one`: {}, `One`: {}}.Values(`one`))
	eq(t, []string{}, H{`one`: {}}.Values(`one`))
	eq(t, []string{}, H{`One`: {}}.Values(`one`))
	eq(t, []string{`two`}, H{`one`: {`two`}, `One`: {`three`}}.Values(`one`))
	eq(t, []string{`three`}, H{`One`: {`three`}}.Values(`one`))
	eq(t, []string{`three`}, H{`one`: {`two`}, `One`: {`three`}}.Values(`One`))
}

func TestHead_Has(t *testing.T) {
	type H = gr.Head

	eq(t, false, gr.Head(nil).Has(`one`))
	eq(t, false, gr.Head(nil).Has(`One`))
	eq(t, true, H{`one`: nil, `One`: nil}.Has(`one`))
	eq(t, true, H{`one`: nil}.Has(`one`))
	eq(t, true, H{`One`: nil}.Has(`one`))
	eq(t, true, H{`one`: nil, `One`: nil}.Has(`One`))
	eq(t, false, H{`one`: nil}.Has(`One`))
	eq(t, true, H{`One`: nil}.Has(`One`))
}

func TestHead_Del(t *testing.T) {
	type H = gr.Head

	test := func(exp, src gr.Head, key string) {
		t.Helper()
		tar := src.Del(key)
		eq(t, exp, tar)
		is(t, src, tar)
	}

	test(nil, nil, `one`)
	test(H{}, H{}, `one`)
	test(H{`two`: {`three`}}, H{`two`: {`three`}}, `one`)
	test(H{}, H{`one`: nil, `One`: nil}, `one`)
	test(H{}, H{`one`: nil}, `one`)
	test(H{}, H{`One`: nil}, `one`)
	test(H{`one`: nil}, H{`one`: nil, `One`: nil}, `One`)
	test(H{`one`: nil}, H{`one`: nil}, `One`)
	test(H{}, H{`One`: nil}, `One`)
	test(H{`two`: {`three`}}, H{`one`: nil, `One`: nil, `two`: {`three`}}, `one`)
	test(H{`two`: {`three`}}, H{`one`: nil, `two`: {`three`}}, `one`)
	test(H{`two`: {`three`}}, H{`One`: nil, `two`: {`three`}}, `one`)
	test(H{`one`: nil, `two`: {`three`}}, H{`one`: nil, `One`: nil, `two`: {`three`}}, `One`)
	test(H{`one`: nil, `two`: {`three`}}, H{`one`: nil, `two`: {`three`}}, `One`)
	test(H{`two`: {`three`}}, H{`One`: nil, `two`: {`three`}}, `One`)
}

func TestHead_Set(t *testing.T) {
	type H = gr.Head

	test := func(exp, src gr.Head, key, val string) {
		t.Helper()

		tar := src.Set(key, val)
		eq(t, exp, tar)

		if src != nil {
			is(t, src, tar)
		}
	}

	test(H{``: {``}}, nil, ``, ``)
	test(H{`One`: {``}}, nil, `one`, ``)
	test(H{`One`: {``}}, nil, `One`, ``)
	test(H{`one`: {`two`}, `One`: {``}}, H{`one`: {`two`}, `One`: {`three`}}, `One`, ``)
	test(H{`one`: {`two`}, `One`: {``}}, H{`one`: {`two`}}, `One`, ``)
	test(H{`One`: {``}}, H{`One`: {`two`}}, `One`, ``)

	test(H{`One`: {`two`}}, nil, `one`, `two`)
	test(H{`One`: {`two`}}, nil, `One`, `two`)
	test(H{`one`: {`two`}, `One`: {`four`}}, H{`one`: {`two`}, `One`: {`three`}}, `One`, `four`)
	test(H{`one`: {`two`}, `One`: {`four`}}, H{`one`: {`two`}}, `One`, `four`)
	test(H{`One`: {`four`}}, H{`One`: {`two`}}, `One`, `four`)

	t.Run(`no slice mutation`, func(t *testing.T) {
		srcVals := []string{`two`}
		srcHead := H{`One`: srcVals}
		tarHead := srcHead.Set(`One`, `three`)

		is(t, srcHead, tarHead)
		eq(t, H{`One`: {`three`}}, tarHead)
		eq(t, []string{`two`}, srcVals)
	})
}

func TestHead_Replace(t *testing.T) {
	type H = gr.Head

	test := func(exp, src gr.Head, key string, vals []string) {
		t.Helper()

		tar := src.Replace(key, vals...)
		eq(t, exp, tar)

		if src != nil {
			is(t, src, tar)
		}
	}

	test(nil, nil, ``, nil)
	test(nil, nil, `one`, nil)
	test(nil, nil, `One`, nil)

	test(nil, nil, ``, []string{})
	test(nil, nil, `one`, []string{})
	test(nil, nil, `One`, []string{})

	test(H{`One`: {`two`}}, nil, `one`, []string{`two`})
	test(H{`One`: {`two`}}, nil, `One`, []string{`two`})

	test(H{}, H{`one`: {`two`}}, `one`, nil)
	test(H{}, H{`one`: {`two`}}, `one`, []string{})
	test(H{`one`: {`two`}}, H{`one`: {`two`}}, `One`, nil)
	test(H{`one`: {`two`}}, H{`one`: {`two`}}, `One`, []string{})
	test(H{`three`: {`four`}}, H{`one`: {`two`}, `three`: {`four`}}, `one`, nil)
	test(H{`three`: {`four`}}, H{`one`: {`two`}, `three`: {`four`}}, `one`, []string{})
	test(H{`one`: {`two`}, `three`: {`four`}}, H{`one`: {`two`}, `three`: {`four`}}, `One`, nil)
	test(H{`one`: {`two`}, `three`: {`four`}}, H{`one`: {`two`}, `three`: {`four`}}, `One`, []string{})

	test(
		H{`One`: {`three`, `four`}},
		H{`one`: {`two`}},
		`one`,
		[]string{`three`, `four`},
	)

	test(
		H{`one`: {`two`}, `One`: {`three`, `four`}},
		H{`one`: {`two`}},
		`One`,
		[]string{`three`, `four`},
	)

	test(
		H{`three`: {`four`}, `One`: {`five`, `six`}},
		H{`one`: {`two`}, `three`: {`four`}},
		`one`,
		[]string{`five`, `six`},
	)

	test(
		H{`one`: {`two`}, `three`: {`four`}, `One`: {`five`, `six`}},
		H{`one`: {`two`}, `three`: {`four`}},
		`One`,
		[]string{`five`, `six`},
	)

	t.Run(`no slice mutation`, func(t *testing.T) {
		srcVals := []string{`two`}
		srcHead := H{`One`: srcVals}
		tarHead := srcHead.Replace(`One`, `three`)

		is(t, srcHead, tarHead)
		eq(t, H{`One`: {`three`}}, tarHead)
		eq(t, []string{`two`}, srcVals)
	})
}

func TestHead_Patch(t *testing.T) {
	type H = gr.Head

	test := func(exp, src H, patch H) {
		t.Helper()

		tar := src.Patch(patch)
		eq(t, exp, tar)

		if src != nil {
			is(t, src, tar)
		}
	}

	test(nil, nil, nil)
	test(nil, nil, H{})
	test(H{}, H{}, nil)
	test(H{}, H{}, H{})

	test(
		H{`One`: {`three`}},
		H{`one`: {`two`}},
		H{`one`: {`three`}},
	)

	test(
		H{`One`: {`five`}, `three`: {`four`}},
		H{`one`: {`two`}, `three`: {`four`}},
		H{`one`: {`five`}},
	)

	test(
		H{`One`: {`five`}, `Three`: {`six`, `seven`}},
		H{`one`: {`two`}, `three`: {`four`}},
		H{`one`: {`five`}, `three`: {`six`, `seven`}},
	)
}
