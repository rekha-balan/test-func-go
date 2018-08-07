package runtime

import (
	"reflect"

	"github.com/Azure/azure-functions-go/internal/rpc"
)

// function contains a function symbol with in and out param types
type function struct {
	handler   reflect.Value
	signature reflect.Type
	in        map[string]*funcField
	out       map[string]*funcField
}

// funcField represents a representation of a func field
type funcField struct {
	Name     string
	Type     reflect.Type
	Binding  *rpc.BindingInfo
	Position int
}

//Call executes the binded function and returns the output
func (f *function) Invoke(params []reflect.Value) ([]reflect.Value, error) {
	output := f.handler.Call(params)
	return output, nil
}
