package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Azure/azure-functions-go/azfunc"
)

// Run runs this Azure Function because it is specified in `function.json` as
// the entryPoint. Fields of the function's parameters are also bound to
// incoming and outgoing event properties as specified in `function.json`.
func Run(ctx azfunc.Context, req *http.Request) (resp *http.Response, err error) {

	// additional properties are bound to ctx by Azure Functions
	ctx.Log(azfunc.LogInformation, "function invoked: function %v, invocation %v",
		ctx.FunctionID(), ctx.InvocationID())

	// get query param values
	name := req.URL.Query().Get("name")
	respBody := fmt.Sprintf("Azure Functions for Go says: Hello world, %s!", name)

	resp = &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          ioutil.NopCloser(bytes.NewBufferString(respBody)),
		ContentLength: int64(len(respBody)),
		Request:       req,
		Header:        make(http.Header, 0),
	}

	if name == "" {
		resp.StatusCode = 400
		resp.Status = "400 BAD REQUEST"
		err = fmt.Errorf("Missing required parameter: name")
	}

	resp.Header.Add("customHeader", "azfuncHello")
	return
}
