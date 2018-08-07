package main

import (
	"net/http"

	"github.com/Azure/azure-functions-go/azfunc"
)

// Run is the entrypoint to our Go Azure Function - if you want to change it, see function.json
func Run(ctx azfunc.Context, req *http.Request, in map[string]interface{}) (out map[string]interface{}) {
	ctx.Log(azfunc.LogInformation, "function id: %s, invocation id: %s with person name: %v", ctx.FunctionID(), ctx.InvocationID(), in["name"])

	out = map[string]interface{}{}
	out["name"] = "new name"
	out["RowKey"] = "newTestKey"
	return
}
