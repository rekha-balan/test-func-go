package main

import (
	"io/ioutil"
	"net/http"

	"github.com/Azure/azure-functions-go/azfunc"
)

// Run is the entrypoint to our Go Azure Function - if you want to change it, see function.json
func Run(ctx azfunc.Context, req *http.Request) (out string) {
	ctx.Log(azfunc.LogInformation, "Log message from function %v, invocation %v to the runtime", ctx.FunctionID(), ctx.InvocationID())

	body, _ := ioutil.ReadAll(req.Body)
	out = string(body)
	return
}
