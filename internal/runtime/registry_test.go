// +build linux darwin
// +build go1.10
// +build cgo

package runtime

import (
	"bytes"
	"net/http"
	"reflect"
	"testing"

	"github.com/vladbarosan/test-func-go/azfunc"
	"github.com/vladbarosan/test-func-go/internal/rpc"
	"github.com/golang/protobuf/jsonpb"
)

func TestLoadFunc_HttpTriggerBlobBindings(t *testing.T) {
	lr := loadFunctionLoadRequest(t, "httpTriggerBlobBindings_FunctionLoadRequest.json")

	r := NewRegistry()
	err := r.LoadFunc(lr)
	if err != nil {
		t.Fatalf("failed to get a function, got error: %v", err)
	}
	f := r.funcs[lr.FunctionId]
	if got, want := f.signature, reflect.TypeOf(func(azfunc.Context, *http.Request, *string) string { return "" }); got != want {
		t.Logf("got:  %q\nwant: %q", got, want)
		t.Fail()
	}
	if got, want := f.out["outBlob"].Name, "outBlob"; got != want {
		t.Logf("got:  %q\nwant: %q", got, want)
		t.Fail()
	}
	if got, want := f.in["req"].Name, "req"; got != want {
		t.Logf("got:  %q\nwant: %q", got, want)
		t.Fail()
	}
	if got, want := f.in["inBlob"].Name, "inBlob"; got != want {
		t.Logf("got:  %q\nwant: %q", got, want)
		t.Fail()
	}
}

func loadFunctionLoadRequest(t *testing.T, name string) *rpc.FunctionLoadRequest {
	b := loadTestData(t, name)
	r := bytes.NewReader(b)
	var lr rpc.FunctionLoadRequest

	jsonpb.Unmarshal(r, &lr)
	return &lr
}
