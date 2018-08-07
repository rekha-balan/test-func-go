package main

import (
	"fmt"

	"github.com/Azure/azure-functions-go/azfunc"
)

// Run is the entrypoint to our Go Azure Function - if you want to change it, see function.json
func Run(ctx azfunc.Context, ehMsg *azfunc.EventHubEvent) (outMsgs []string) {
	ctx.Log(azfunc.LogInformation, "Log message from function %v, invocation %v to the runtime with batch data: %s and msg %v", ctx.FunctionID(), ctx.InvocationID(), ehMsg)

	outMsgs = make([]string, 10)
	for i := 0; i < 10; i++ {
		outMsgs[i] = fmt.Sprintf("Message %d from Azure Functions for Go", i)
	}
	return
}
