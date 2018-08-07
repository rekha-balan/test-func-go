package main

import (
	"fmt"

	"github.com/Azure/azure-functions-go/azfunc"
)

// Run is the entrypoint to our Go Azure Function - if you want to change it, see function.json
func Run(ctx azfunc.Context, msg *azfunc.SBMsg) (task *Task) {
	ctx.Log(azfunc.LogInformation, "Log message from function %v, invocation %v", ctx.FunctionID(), ctx.InvocationID())

	ctx.Log(azfunc.LogInformation, "Creating new task from %s with priority %d", msg.Data, msg.DeliveryCount)
	task = &Task{
		Name:     fmt.Sprintf("%d-task", msg.DeliveryCount),
		Priority: 1,
		Type:     "investigation",
	}

	return
}

// Task represents work that needs to be done.
type Task struct {
	Name     string
	Priority int
	Type     string
}

// Report represents written content by an author.
type Report struct {
	ID       string
	Priority int
	Format   string
	Author   string
	Content  string
}
