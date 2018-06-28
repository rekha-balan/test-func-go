package azfunc

import (
	"net/http"
	"reflect"

	"github.com/Azure/azure-functions-go-worker/logger"
	"github.com/Azure/azure-functions-go-worker/rpc"
)

// TriggerType represents the supported trigger types.
type TriggerType string

// BindingType represents the supported binding types.
type BindingType string

const (
	// HTTPTrigger represents a HTTP trigger in function load request from the host
	HTTPTrigger BindingType = "httpTrigger"

	// BlobTrigger represents a blob trigger in function load request from host
	BlobTrigger BindingType = "blobTrigger"

	// HTTPBinding represents a HTTP binding in function load request from the host
	HTTPBinding BindingType = "http"

	// BlobBinding represents a blob binding in function load request from the host
	BlobBinding BindingType = "blob"
)

// StringToType - Because we don't have go/types information, we need to map the type info from the AST (which is string) to the actual types - see loader.go:83
// investiage automatically adding here all types from package azfunc
var StringToType = map[string]reflect.Type{
	"*http.Request":   reflect.TypeOf((*http.Request)(nil)),
	"*azfunc.Context": reflect.TypeOf((*Context)(nil)),
	"*azfunc.Blob":    reflect.TypeOf((*Blob)(nil)),
}

// Func contains a function symbol with in and out param types
type Func struct {
	Func             reflect.Value
	Bindings         map[string]*rpc.BindingInfo
	In               []reflect.Type
	NamedInArgs      []*Arg
	Out              []reflect.Type
	NamedOutBindings map[string]reflect.Value
}

// Context contains the runtime context of the function
type Context struct {
	FunctionID   string
	InvocationID string
	Logger       *logger.Logger
}

// Arg represents an initial representation of a func argument
type Arg struct {
	Name string
	Type reflect.Type
}
