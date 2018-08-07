package main

import (
	"github.com/Azure/azure-functions-go/azfunc"
)

// Run is the entrypoint to our Go Azure Function - if you want to change it, see function.json
func Run(ctx azfunc.Context, event *azfunc.EventGridEvent) {
	ctx.Log(azfunc.LogInformation, "Log message from function %v, invocation %v to the runtime is from topic: %v", ctx.FunctionID(), ctx.InvocationID(), event.Topic)
}
