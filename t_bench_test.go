package gr_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/textproto"
	"strings"
	"testing"

	"github.com/mitranim/gr"
)

var (
	reqNop          = func(*Q) {}
	ioReadCloserNop = func(io.ReadCloser) {}
	stringNop       = func(string) {}
	stringsNop      = func([]string) {}
)

func Benchmark_req_chain(b *testing.B) {
	for range iter(b.N) {
		reqNop(
			new(gr.Req).
				Path(`/one`).
				RawQuery(`two=three`).
				Post().
				TypeJson().
				String(`hello world`).
				Init().
				Req(),
		)
	}
}

func Benchmark_req_std(b *testing.B) {
	for range iter(b.N) {
		req, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodPost,
			`/one`,
			strings.NewReader(`hello world`),
		)
		try(err)
		reqNop(req)
	}
}

func Benchmark_req_gr(b *testing.B) {
	for range iter(b.N) {
		reqNop(
			gr.Ctx(context.Background()).
				Path(`/one`).
				Post().
				String(`hello world`).
				Init().
				Req(),
		)
	}
}

func Benchmark_io_NopCloser_strings_NewReader(b *testing.B) {
	for range iter(b.N) {
		ioReadCloserNop(io.NopCloser(strings.NewReader(`hello world`)))
	}
}

func BenchmarkNewStringReadCloser(b *testing.B) {
	for range iter(b.N) {
		ioReadCloserNop(gr.NewStringReadCloser(`hello world`))
	}
}

func Benchmark_io_NopCloser_bytes_NewReader(b *testing.B) {
	src := []byte(`hello world`)
	b.ResetTimer()

	for range iter(b.N) {
		ioReadCloserNop(io.NopCloser(bytes.NewReader(src)))
	}
}

func BenchmarkNewBytesReadCloser(b *testing.B) {
	src := []byte(`hello world`)
	b.ResetTimer()

	for range iter(b.N) {
		ioReadCloserNop(gr.NewBytesReadCloser(src))
	}
}

func Benchmark_textproto_CanonicalMIMEHeaderKey_common_non_canon(b *testing.B) {
	for range iter(b.N) {
		stringNop(textproto.CanonicalMIMEHeaderKey(`cOnTeNt-TyPe`))
	}
}

func Benchmark_textproto_CanonicalMIMEHeaderKey_custom_non_canon(b *testing.B) {
	for range iter(b.N) {
		stringNop(textproto.CanonicalMIMEHeaderKey(`x-one-two-three`))
	}
}

func Benchmark_textproto_CanonicalMIMEHeaderKey_custom_canon(b *testing.B) {
	for range iter(b.N) {
		stringNop(textproto.CanonicalMIMEHeaderKey(`X-One-Two-Three`))
	}
}

func Benchmark_header_values_miss(b *testing.B) {
	for range iter(b.N) {
		stringsNop(H{
			`one-two-three`: {`four-five`, `six-seven`},
			`One-Two-Three`: {`four-five`, `six-seven`},
		}.Values(`one-two-three`))
	}
}

func Benchmark_header_values_hit(b *testing.B) {
	for range iter(b.N) {
		stringsNop(H{
			`one-two-three`: {`four-five`, `six-seven`},
			`One-Two-Three`: {`four-five`, `six-seven`},
		}.Values(`One-Two-Three`))
	}
}

func Benchmark_head_values_non_canon_hit(b *testing.B) {
	for range iter(b.N) {
		stringsNop(gr.Head{
			`one-two-three`: {`four-five`, `six-seven`},
			`One-Two-Three`: {`four-five`, `six-seven`},
		}.Values(`one-two-three`))
	}
}

func Benchmark_head_values_canon_hit(b *testing.B) {
	for range iter(b.N) {
		stringsNop(gr.Head{
			`one-two-three`: {`four-five`, `six-seven`},
			`One-Two-Three`: {`four-five`, `six-seven`},
		}.Values(`One-Two-Three`))
	}
}

func Benchmark_header_del_miss(b *testing.B) {
	for range iter(b.N) {
		H{
			`one-two-three`: {`four-five`, `six-seven`},
			`One-Two-Three`: {`four-five`, `six-seven`},
		}.Del(`one-two-three`)
	}
}

func Benchmark_header_del_hit(b *testing.B) {
	for range iter(b.N) {
		H{
			`one-two-three`: {`four-five`, `six-seven`},
			`One-Two-Three`: {`four-five`, `six-seven`},
		}.Del(`One-Two-Three`)
	}
}

func Benchmark_head_del_non_canon_hit(b *testing.B) {
	for range iter(b.N) {
		gr.Head{
			`one-two-three`: {`four-five`, `six-seven`},
			`One-Two-Three`: {`four-five`, `six-seven`},
		}.Del(`one-two-three`)
	}
}

func Benchmark_head_del_canon_hit(b *testing.B) {
	for range iter(b.N) {
		gr.Head{
			`one-two-three`: {`four-five`, `six-seven`},
			`One-Two-Three`: {`four-five`, `six-seven`},
		}.Del(`One-Two-Three`)
	}
}

func Benchmark_header_set_miss(b *testing.B) {
	for range iter(b.N) {
		H{
			`one-two-three`: {`four-five`, `six-seven`},
			`One-Two-Three`: {`four-five`, `six-seven`},
		}.Set(`one-two-three`, `eight-nine`)
	}
}

func Benchmark_header_set_hit(b *testing.B) {
	for range iter(b.N) {
		H{
			`one-two-three`: {`four-five`, `six-seven`},
			`One-Two-Three`: {`four-five`, `six-seven`},
		}.Set(`One-Two-Three`, `eight-nine`)
	}
}

func Benchmark_head_set_non_canon_hit(b *testing.B) {
	for range iter(b.N) {
		gr.Head{
			`one-two-three`: {`four-five`, `six-seven`},
			`One-Two-Three`: {`four-five`, `six-seven`},
		}.Set(`one-two-three`, `eight-nine`)
	}
}

func Benchmark_head_set_canon_hit(b *testing.B) {
	for range iter(b.N) {
		gr.Head{
			`one-two-three`: {`four-five`, `six-seven`},
			`One-Two-Three`: {`four-five`, `six-seven`},
		}.Set(`One-Two-Three`, `eight-nine`)
	}
}
