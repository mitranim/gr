package gr_test

import (
	"fmt"
	"net/url"

	"github.com/mitranim/gr"
)

type Input struct {
	InputVal string `json:"inputVal"`
}

type Output struct {
	ReqMethod string `json:"reqMethod"`
	ReqUrl    string `json:"reqUrl"`
	ReqBody   string `json:"reqBody"`
}

func ExampleReq_jsonInputJsonOutput() {
	input := Input{`some input`}

	var output Output
	gr.To(testServer.URL).Path(`/json`).Json(input).Res().Ok().Json(&output)

	fmt.Printf("%#v\n", output)

	// Output:
	// gr_test.Output{ReqMethod:"GET", ReqUrl:"/json", ReqBody:"{\"inputVal\":\"some input\"}"}
}

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
