package gr_test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	ht "net/http/httptest"
	"net/url"
	"os"
	r "reflect"
	"runtime"
	"strings"
	"testing"
	"unsafe"
)

type (
	U    = url.URL
	V    = url.Values
	H    = http.Header
	Q    = http.Request
	S    = http.Response
	W    = http.ResponseWriter
	Pair = [2]interface{}
)

var testServer *ht.Server

func TestMain(m *testing.M) {
	testServer = ht.NewServer(http.HandlerFunc(testHandler))
	defer testServer.Close()
	os.Exit(m.Run())
}

func testHandler(rew W, req *Q) {
	switch req.URL.Path {
	case `/json`:
		testHandlerJson(rew, req)
	default:
		testHandlerDefault(rew, req)
	}
}

func testHandlerJson(rew W, req *Q) {
	try(json.NewEncoder(rew).Encode(struct {
		ReqMethod string `json:"reqMethod"`
		ReqUrl    string `json:"reqUrl"`
		ReqBody   string `json:"reqBody"`
	}{
		ReqMethod: req.Method,
		ReqUrl:    req.URL.String(),
		ReqBody:   readStr(req.Body),
	}))
}

func testHandlerDefault(rew W, req *Q) {
	fmt.Fprintf(rew, `
request method: %v
request URL: %v
request body: %s
`, req.Method, req.URL, readStr(req.Body))
}

func eq(t testing.TB, exp, act interface{}) {
	t.Helper()
	if !r.DeepEqual(exp, act) {
		t.Fatalf(`
expected (detailed):
	%#[1]v
actual (detailed):
	%#[2]v
expected (simple):
	%[1]v
actual (simple):
	%[2]v
`, exp, act)
	}
}

func is(t testing.TB, exp, act interface{}) {
	t.Helper()

	// nolint:structcheck
	type iface struct {
		typ unsafe.Pointer
		dat unsafe.Pointer
	}

	expIface := *(*iface)(unsafe.Pointer(&exp))
	actIface := *(*iface)(unsafe.Pointer(&act))

	if expIface != actIface {
		t.Fatalf(`
expected (interface):
	%#[1]v
actual (interface):
	%#[2]v
expected (detailed):
	%#[3]v
actual (detailed):
	%#[4]v
expected (simple):
	%[3]v
actual (simple):
	%[4]v
`, expIface, actIface, exp, act)
	}
}

func errs(t testing.TB, msg string, err error) {
	if err == nil {
		t.Fatalf(`expected an error with %q, got none`, msg)
	}

	str := err.Error()
	if !strings.Contains(str, msg) {
		t.Fatalf(`expected an error with a message containing %q, got %q`, msg, str)
	}
}

func panics(t testing.TB, msg string, fun func()) {
	t.Helper()
	val := catchAny(fun)

	if val == nil {
		t.Fatalf(`expected %v to panic, found no panic`, funcName(fun))
	}

	str := fmt.Sprint(val)
	if !strings.Contains(str, msg) {
		t.Fatalf(`
expected %v to panic with a message containing:
	%v
found the following message:
	%v
`, funcName(fun), msg, str)
	}
}

func eqTest(t testing.TB) func(interface{}, interface{}) {
	return func(exp, act interface{}) {
		t.Helper()
		eq(t, exp, act)
	}
}

func codeTest(t testing.TB, fun func(int) bool) func(bool, int) {
	return func(exp bool, code int) {
		t.Helper()
		eq(t, exp, fun(code))
	}
}

func funcName(val interface{}) string {
	return runtime.FuncForPC(r.ValueOf(val).Pointer()).Name()
}

func catchAny(fun func()) (val interface{}) {
	defer recAny(&val)
	fun()
	return
}

func recAny(ptr *interface{}) { *ptr = recover() }

func try(err error) {
	if err != nil {
		panic(err)
	}
}

type Trans struct {
	Req *http.Request
	Res *http.Response
	Err error
}

func (self *Trans) RoundTrip(req *http.Request) (*http.Response, error) {
	self.Req = req
	return self.Res, self.Err
}

var (
	errRead   = fmt.Errorf(`unexpected read`)
	errClose  = fmt.Errorf(`unexpected close`)
	errDecode = fmt.Errorf(`failing to decode`)
)

type FailReadCloser struct{}

func (FailReadCloser) Read([]byte) (int, error) { return 0, errRead }
func (FailReadCloser) Close() error             { return errClose }

type ReaderFailCloser struct{}

func (ReaderFailCloser) Read([]byte) (int, error) { return 0, io.EOF }
func (ReaderFailCloser) Close() error             { return errClose }

type ReadCloseFlag struct {
	DidRead  bool
	DidClose bool
}

func (self *ReadCloseFlag) Read([]byte) (int, error) {
	self.DidRead = true
	return 0, io.EOF
}

func (self *ReadCloseFlag) Close() error {
	if self.DidClose {
		return errClose
	}
	self.DidClose = true
	return nil
}

func NewReaderCloseFlag(val string) *ReaderCloseFlag {
	return &ReaderCloseFlag{Reader: strings.NewReader(val)}
}

type ReaderCloseFlag struct {
	io.Reader
	DidClose bool
}

func (self *ReaderCloseFlag) Close() error {
	if self.DidClose {
		return errClose
	}
	self.DidClose = true
	return nil
}

type DecodeFail struct{}

func (self *DecodeFail) UnmarshalJSON([]byte) error                        { return errDecode }
func (self *DecodeFail) UnmarshalXML(*xml.Decoder, xml.StartElement) error { return errDecode }

func spr(val string) *string { return &val }

func iter(count int) []struct{} { return make([]struct{}, count) }

func readStr(src io.Reader) string {
	if src == nil {
		return ``
	}

	val, err := io.ReadAll(src)
	try(err)

	return string(val)
}
