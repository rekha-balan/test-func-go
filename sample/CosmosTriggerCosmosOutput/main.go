package main

import (
	"fmt"

	"github.com/Azure/azure-functions-go/azfunc"
)

// Run is the entrypoint to our Go Azure Function - if you want to change it, see function.json
func Run(ctx azfunc.Context, reports []Report) (tasks []Task) {
	ctx.Log(azfunc.LogInformation, "Log message from function %v, invocation %v with %d reports", ctx.FunctionID(), ctx.InvocationID(), len(reports))

	for _, report := range reports {
		ctx.Log(azfunc.LogInformation, "Creating new task from %s with priority %d", report.ID, report.Priority)
		t := Task{
			Name:     fmt.Sprintf("%s-task", report.ID),
			Priority: report.Priority,
			Type:     "investigation",
		}
		tasks = append(tasks, t)
	}

	ctx.Log(azfunc.LogInformation, "Created %d tasks", len(tasks))
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
