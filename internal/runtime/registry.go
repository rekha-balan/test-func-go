package runtime

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"plugin"
	"reflect"

	"github.com/Azure/azure-functions-go/internal/rpc"
	logrus "github.com/Sirupsen/logrus"
)

// Registry contains all information about user functions and how to execute them
type Registry struct {
	funcs map[string]*function
}

// NewRegistry returns a new function registry
func NewRegistry() *Registry {
	return &Registry{
		funcs: map[string]*function{},
	}
}

// LoadFunc populates information about the func from the compiled plugin and from parsing the source code
func (r Registry) LoadFunc(req *rpc.FunctionLoadRequest) error {
	logrus.Debugf("received function load request: %v", req)

	f, err := loadFuncFromPlugin(req.Metadata)
	if err != nil {
		return fmt.Errorf("cannot load function from plugin: %v", err)
	}

	ins, outs, err := loadInOut(req.Metadata, f.signature)
	if err != nil {
		return fmt.Errorf("cannot parse entrypoint: %v", err)
	}

	f.in = ins
	f.out = outs

	logrus.Debugf("function: %v", f)
	r.funcs[req.FunctionId] = f

	return nil
}

// ExecuteFunc takes an InvocationRequest and executes the function with corresponding function ID
func (r Registry) ExecuteFunc(req *rpc.InvocationRequest, eventStream rpc.FunctionRpc_EventStreamClient) (response *rpc.InvocationResponse) {

	logrus.Debugf("\n\n\nInvocation Request: %v", req)

	status := rpc.StatusResult_Success

	ir := &rpc.InvocationResponse{
		InvocationId: req.InvocationId,
		Result: &rpc.StatusResult{
			Status: status,
		},
	}
	f, ok := r.funcs[req.FunctionId]

	if !ok {
		logrus.Debugf("function with functionID %v not loaded", req.FunctionId)
		ir.Result.Status = rpc.StatusResult_Failure
		return ir
	}

	args, err := FromProto(req, f.in)
	if err != nil {
		ir.Result.Status = rpc.StatusResult_Failure
		return ir
	}

	params := make([]reflect.Value, len(f.in))
	for _, v := range f.in {
		isIntf := v.Type.Kind() == reflect.Interface
		logrus.Debugf("Kind is %v  and is intf:%t and type is %v", v.Type.Kind(), isIntf, v.Type)
		ctx := &funcContext{
			Context:      context.Background(),
			functionID:   req.FunctionId,
			invocationID: req.InvocationId,
			eventStream:  eventStream,
		}
		ctxv := reflect.ValueOf(ctx).Elem()

		if v.Type.Kind() == reflect.Interface && ctxv.Type().Implements(v.Type) {
			logrus.Debug("created context")
			params[v.Position] = ctxv
		} else {
			params[v.Position] = args[v.Name]
		}
	}

	output, err := f.Invoke(params)
	if err != nil {
		ir.Result.Status = rpc.StatusResult_Failure
		return ir
	}
	o, rv, s, err := ToProto(output, f.out)

	if err != nil {
		logrus.Debugf("cannot get output data from result %v", err)
		if err != nil {
			ir.Result.Status = rpc.StatusResult_Failure
			return ir
		}
	}

	ir.ReturnValue = rv
	ir.OutputData = o
	ir.Result = s
	return ir
}

// funcContext implements the azfunc.Context interface
type funcContext struct {
	context.Context
	functionID   string
	invocationID string
	eventStream  rpc.FunctionRpc_EventStreamClient
}

func (c funcContext) FunctionID() string {
	return c.functionID
}

func (c funcContext) InvocationID() string {
	return c.invocationID
}

func (c funcContext) Log(level int, format string, args ...interface{}) error {
	rpcLevel := rpc.RpcLog_Level(level)
	if rpcLevel < rpc.RpcLog_Trace {
		rpcLevel = rpc.RpcLog_Trace
	}
	if rpcLevel > rpc.RpcLog_None {
		rpcLevel = rpc.RpcLog_Critical
	}

	l := &rpc.RpcLog{
		InvocationId: c.invocationID,
		Level:        rpcLevel,
		Message:      fmt.Sprintf(format, args...),
	}

	return c.eventStream.Send(&rpc.StreamingMessage{
		Content: &rpc.StreamingMessage_RpcLog{
			RpcLog: l,
		},
	})
}

// loadFuncFromPlugin takes the compiled plugin from the func's bin directory
// then reads through reflection the in and out paramns of the entrypoint
func loadFuncFromPlugin(metadata *rpc.RpcFunctionMetadata) (*function, error) {

	path := fmt.Sprintf("%s/bin/%s.so", metadata.Directory, metadata.Name)
	plugin, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot get .so object from path %s: %v", path, err)
	}

	symbol, err := plugin.Lookup(metadata.EntryPoint)
	if err != nil {
		return nil, fmt.Errorf("cannot look up symbol for entrypoint function %s: %v", metadata.EntryPoint, err)
	}

	t := reflect.TypeOf(symbol)
	if t.Kind() != reflect.Func {
		return nil, fmt.Errorf("symbol is not func, but %v", t.Kind())
	}

	return &function{
		handler:   reflect.ValueOf(symbol),
		signature: t,
	}, nil
}

type iterator func(int) reflect.Type

// loadInOut loads the input and output types for a function
func loadInOut(metadata *rpc.RpcFunctionMetadata, funcType reflect.Type) (map[string]*funcField, map[string]*funcField, error) {

	var ins, outs map[string]*funcField
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, metadata.ScriptFile, nil, parser.AllErrors)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot parse file %v: %v", metadata.ScriptFile, err)
	}

	// traverse the AST and inspect the nodes
	// if the node is a func declaration, check if entrypoint and get input params names and types (as string)
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			logrus.Debugf("found function: %v", x.Name.Name)
			if x.Name.Name != metadata.EntryPoint {
				logrus.Debugf("not function entrypoint, moving on...")

				// not the entrypoint, go further into the AST
				return true
			}

			ins, err = extractFuncFields(x.Type.Params, metadata.GetBindings(), funcType.In, funcType.NumIn())
			outs, err = extractFuncFields(x.Type.Results, metadata.GetBindings(), funcType.Out, funcType.NumOut())

			// this is the entrypoint, no need to traverse the AST any longer
			return false

		default:
			// not a func declaration, need to go further in the AST
			return true
		}
	})

	return ins, outs, nil
}

func extractFuncFields(fl *ast.FieldList, bindings map[string]*rpc.BindingInfo, fi iterator, l int) (map[string]*funcField, error) {
	fields := map[string]*funcField{}

	if fl.NumFields() != l {
		return nil, fmt.Errorf("Plugin %d and source %d nr of arguments are different", fl.NumFields(), l)
	}

	if l == 0 {
		return fields, nil
	}

	for i, p := range fl.List {
		t := fi(i)
		for _, n := range p.Names {
			logrus.Debugf("Found parameter: %s with type: %s", n, t.String())

			fields[n.Name] = &funcField{
				Name:     n.Name,
				Type:     t,
				Position: i,
				Binding:  bindings[n.Name],
			}
		}
	}

	return fields, nil
}
