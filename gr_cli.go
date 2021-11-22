package gr

import "net/http"

/*
Short for "client" (not "CLI"). Alias of `http.Client` with added shortcuts for
better compatibility with `gr.Req` and `gr.Res`. Freely castable to and from
`http.Client`.
*/
type Cli http.Client

// Free cast to `*http.Client`.
func (self *Cli) Cli() *http.Client { return (*http.Client)(self) }

/*
Returns a new `gr.Req` with this client. Executing the request via
`(*gr.Req).Res` or `(*gr.Req).ResCatch` will use this client.
*/
func (self *Cli) Req() *Req { return new(Req).Cli(self.Cli()) }

/*
Similar to `(*http.Client).Do`, but with "gr" types. Panics on error. To catch
the error, use `(*gr.Cli).DoCatch`.
*/
func (self *Cli) Do(req *Req) *Res { return req.CliRes(self.Cli()) }

// Similar to `(*http.Client).Do`, but with "gr" types.
func (self *Cli) DoCatch(req *Req) (*Res, error) {
	return req.CliResCatch(self.Cli())
}
