package main

import (
	"github.com/Azure/azure-functions-go/azfunc"
)

// Run is the entrypoint to our Go Azure Function - if you want to change it, see function.json
func Run(ctx azfunc.Context, blob *azfunc.Blob) {
	ctx.Log(azfunc.LogInformation, "function id: %s, invocation id: %s with blob : %v", ctx.FunctionID(), ctx.InvocationID(), *blob)
}
