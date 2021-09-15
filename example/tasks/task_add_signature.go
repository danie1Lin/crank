package tasks

import (
	"github.com/RichardKnop/machinery/v1/tasks"
)

func NewTaskAddSignature(a int, b int) *tasks.Signature {
	args := []tasks.Arg{
		{Type: "int", Value: a},
		{Type: "int", Value: b},
	}
	return &tasks.Signature{
		Name: "TaskAdd",
		Args: args,
	}
}
