package gr

import "strconv"

/*
Wraps another error, adding an HTTP status code and response body. Some errors
returned by this package have codes obtained from `http.Response`.
*/
type Err struct {
	Status int    `json:"status,omitempty"`
	Body   []byte `json:"body,omitempty"`
	Cause  error  `json:"cause,omitempty"`
}

// Implement a hidden interface in "errors".
func (self Err) Unwrap() error { return self.Cause }

// Returns `.Status`. Implements a hidden interface supported by
// `github.com/mitranim/rout`.
func (self Err) HttpStatusCode() int { return self.Status }

// Implement the `error` interface.
func (self Err) Error() string {
	return bytesString(self.Append(nil))
}

// Appends the error representation. Used internally by `.Error`.
func (self Err) Append(buf []byte) []byte {
	buf = growBytes(buf, 128)
	buf = append(buf, `[gr] error`...)

	if self.Status != 0 {
		buf = append(buf, ` (HTTP status `...)
		buf = strconv.AppendInt(buf, int64(self.Status), 10)
		buf = append(buf, `)`...)
	}

	cause := self.Cause
	if cause != nil {
		buf = append(buf, `: `...)
		impl, _ := cause.(interface{ Append([]byte) []byte })
		if impl != nil {
			buf = impl.Append(buf)
		} else {
			buf = append(buf, cause.Error()...)
		}
	}

	buf = appendBodyPreview(buf, self.Body)
	return buf
}
