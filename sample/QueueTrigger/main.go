package main

import (
	"github.com/Azure/azure-functions-go-worker/azfunc"
)

// Run is the entrypoint to our Go Azure Function - if you want to change it, see function.json
func Run(ctx azfunc.Context, queueMsg *azfunc.QueueMsg) {
	ctx.Logger.Log("function id: %s, invocation id: %s with queue : %v", ctx.FunctionID, ctx.InvocationID, *queueMsg)
}
