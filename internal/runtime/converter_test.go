package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/Azure/azure-functions-go/azfunc"
	"github.com/Azure/azure-functions-go/internal/rpc"
	"github.com/go-test/deep"
	"github.com/golang/protobuf/jsonpb"
)

func TestConvertToTypeValue_HttpRequest(t *testing.T) {
	ir := loadInvocationRequest(t, "httpTrigger_InvocationRequest.json")

	testCases := []struct {
		want reflect.Type
	}{
		{reflect.TypeOf((*http.Request)(nil))},
		{reflect.TypeOf(http.Request{})},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.want), func(t *testing.T) {
			r, err := convertToTypeValue(tc.want, ir.InputData[0].GetData(), ir.GetTriggerMetadata())

			if err != nil {
				t.Fatalf("failed to get a value, got error: %v", err)
			}

			if got := r.Type(); got != tc.want {
				t.Logf("got:  %q\nwant: %q", got, tc.want)
				t.Fail()
			}

			var v http.Request
			if r.Kind() == reflect.Ptr {
				v = r.Elem().Interface().(http.Request)
			} else {
				v = r.Interface().(http.Request)
			}

			if got, want := v.URL.Query().Get("name"), "testuser"; got != want {
				t.Logf("got:  %q\nwant: %q", got, want)
				t.Fail()
			}

			body, _ := ioutil.ReadAll(v.Body)
			var data map[string]interface{}
			_ = json.Unmarshal(body, &data)

			if got, want := data["password"].(string), "secretPassword"; got != want {
				t.Logf("got:  %q\nwant: %q", got, want)
				t.Fail()
			}
		})
	}
}

func TestConvertToTypeValue_Map(t *testing.T) {

	ir := loadInvocationRequest(t, "tableInput_InvocationRequest.json")

	want := reflect.TypeOf(map[string]interface{}{})
	r, err := convertToTypeValue(want, ir.InputData[0].GetData(), ir.GetTriggerMetadata())

	if err != nil {
		t.Fatalf("failed to get a value, got error: %v", err)
	}

	if got := r.Type(); got != want {
		t.Logf("got:  %q\nwant: %q", got, want)
		t.Fail()
	}

	v := r.Interface().(map[string]interface{})

	if got, want := v["name"], "bestnametest"; got != want {
		t.Logf("got:  %s\nwant: %s", got, want)
		t.Fail()
	}
}

func TestConvertToTypeValue_String(t *testing.T) {
	ir := loadInvocationRequest(t, "blobInput_InvocationRequest.json")
	testCases := []struct {
		want reflect.Type
	}{
		{reflect.TypeOf((*string)(nil))},
		{reflect.TypeOf("")},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.want), func(t *testing.T) {
			r, err := convertToTypeValue(tc.want, ir.InputData[0].GetData(), ir.GetTriggerMetadata())

			if err != nil {
				t.Fatalf("failed to get a value, got error: %v", err)
			}

			if got := r.Type(); got != tc.want {
				t.Logf("got:  %q\nwant: %q", got, tc.want)
				t.Fail()
			}

			var v string
			if r.Kind() == reflect.Ptr {
				v = r.Elem().Interface().(string)
			} else {
				v = r.Interface().(string)
			}

			if got, want := v, "sample input blob content"; got != want {
				t.Logf("got:  %s\nwant: %s", got, want)
				t.Fail()
			}
		})
	}
}

func TestConvertToTypeValue_Timer(t *testing.T) {
	ir := loadInvocationRequest(t, "timerTrigger_InvocationRequest.json")

	testCases := []struct {
		want reflect.Type
	}{
		{reflect.TypeOf((*azfunc.Timer)(nil))},
		{reflect.TypeOf(azfunc.Timer{})},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.want), func(t *testing.T) {
			r, err := convertToTypeValue(tc.want, ir.InputData[0].GetData(), ir.GetTriggerMetadata())

			if err != nil {
				t.Fatalf("failed to get a value, got error: %v", err)
			}

			if got := r.Type(); got != tc.want {
				t.Logf("got:  %q\nwant: %q", got, tc.want)
				t.Fail()
			}

			var v azfunc.Timer
			if r.Kind() == reflect.Ptr {
				v = r.Elem().Interface().(azfunc.Timer)
			} else {
				v = r.Interface().(azfunc.Timer)
			}

			if got, want := v.PastDue, false; got != want {
				t.Logf("got:  %t\nwant: %t", got, want)
				t.Fail()
			}
		})
	}
}

var emptyLocation = time.FixedZone("+0000", 0)

func TestConvertToTypeValue_Blob(t *testing.T) {
	ir := loadInvocationRequest(t, "blobTrigger_InvocationRequest.json")

	testCases := []struct {
		want reflect.Type
	}{
		{reflect.TypeOf((*azfunc.Blob)(nil))},
		{reflect.TypeOf(azfunc.Blob{})},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.want), func(t *testing.T) {
			r, err := convertToTypeValue(tc.want, ir.InputData[0].GetData(), ir.GetTriggerMetadata())

			if err != nil {
				t.Fatalf("failed to get a value, got error: %v", err)
			}

			if got := r.Type(); got != tc.want {
				t.Logf("got:  %q\nwant: %q", got, tc.want)
				t.Fail()
			}

			var v azfunc.Blob
			if r.Kind() == reflect.Ptr {
				v = r.Elem().Interface().(azfunc.Blob)
			} else {
				v = r.Interface().(azfunc.Blob)
			}

			expectedBlob := azfunc.Blob{
				Name:    "testblob.txt",
				Content: "blob content test input",
				URI:     "https://samplestorageaccount.blob.core.windows.net:443/demo/testblob.txt",
				Properties: azfunc.BlobProperties{
					ContentMD5:   "LRhNxuDmIGXy0KzNoxj9bg==",
					ContentType:  "text/plain",
					ETag:         "\"0x8D5EC8302DB81F9\"",
					LastModified: time.Date(2018, time.July, 18, 7, 49, 37, 0, emptyLocation),
					Length:       18,
				},
			}

			got, want := v, expectedBlob
			if diff := deep.Equal(got, want); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func TestConvertToTypeValue_QueueMsg(t *testing.T) {
	ir := loadInvocationRequest(t, "queueMsgTrigger_InvocationRequest.json")

	testCases := []struct {
		want reflect.Type
	}{
		{reflect.TypeOf((*azfunc.QueueMsg)(nil))},
		{reflect.TypeOf(azfunc.QueueMsg{})},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.want), func(t *testing.T) {
			r, err := convertToTypeValue(tc.want, ir.InputData[0].GetData(), ir.GetTriggerMetadata())

			if err != nil {
				t.Fatalf("failed to get a value, got error: %v", err)
			}

			if got := r.Type(); got != tc.want {
				t.Logf("got:  %q\nwant: %q", got, tc.want)
				t.Fail()
			}

			var v azfunc.QueueMsg
			if r.Kind() == reflect.Ptr {
				v = r.Elem().Interface().(azfunc.QueueMsg)
			} else {
				v = r.Interface().(azfunc.QueueMsg)
			}

			expectedQueueMsg := azfunc.QueueMsg{
				ID:           "38c00d86-c30c-4a48-aff5-deafb4b273e4",
				DequeueCount: 1,
				Expiration:   time.Date(2018, time.July, 25, 8, 15, 8, 0, emptyLocation),
				Insertion:    time.Date(2018, time.July, 18, 8, 15, 8, 0, emptyLocation),
				NextVisible:  time.Date(2018, time.July, 18, 8, 25, 15, 0, emptyLocation),
				PopReceipt:   "AgAAAAMAAAAAAAAASsWZ2nAe1AE=",
				Text:         "test queue msg",
			}

			got, want := v, expectedQueueMsg
			if diff := deep.Equal(got, want); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func TestConvertToTypeValue_ServiceBusMsg(t *testing.T) {
	ir := loadInvocationRequest(t, "sbTrigger_InvocationRequest.json")

	testCases := []struct {
		want reflect.Type
	}{
		{reflect.TypeOf((*azfunc.SBMsg)(nil))},
		{reflect.TypeOf(azfunc.SBMsg{})},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.want), func(t *testing.T) {
			r, err := convertToTypeValue(tc.want, ir.InputData[0].GetData(), ir.GetTriggerMetadata())

			if err != nil {
				t.Fatalf("failed to get a value, got error: %v", err)
			}

			if got := r.Type(); got != tc.want {
				t.Logf("got:  %q\nwant: %q", got, tc.want)
				t.Fail()
			}

			var v azfunc.SBMsg
			if r.Kind() == reflect.Ptr {
				v = r.Elem().Interface().(azfunc.SBMsg)
			} else {
				v = r.Interface().(azfunc.SBMsg)
			}

			expectedQueueMsg := azfunc.SBMsg{
				Data:            "Message 1",
				MessageID:       "429c66a736a94a2e8c6e2783e568d460",
				DeliveryCount:   7,
				ExpiresAtUtc:    time.Date(2018, time.July, 31, 23, 54, 18, 288000000, time.UTC),
				EnqueuedTimeUtc: time.Date(2018, 7, 30, 23, 54, 18, 288000000, time.UTC),
				SequenceNumber:  281474976710657,
				UserProperties: map[string]interface{}{
					"x-opt-enqueue-sequence-number": float64(0),
				},
			}

			got, want := v, expectedQueueMsg
			if diff := deep.Equal(got, want); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func TestConvertToTypeValue_EventGridEvent(t *testing.T) {
	ir := loadInvocationRequest(t, "eventGridEventTrigger_InvocationRequest.json")

	testCases := []struct {
		want reflect.Type
	}{
		{reflect.TypeOf((*azfunc.EventGridEvent)(nil))},
		{reflect.TypeOf(azfunc.EventGridEvent{})},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.want), func(t *testing.T) {
			r, err := convertToTypeValue(tc.want, ir.InputData[0].GetData(), ir.GetTriggerMetadata())

			if err != nil {
				t.Fatalf("failed to get a value, got error: %v", err)
			}

			if got := r.Type(); got != tc.want {
				t.Logf("got:  %q\nwant: %q", got, tc.want)
				t.Fail()
			}

			var v azfunc.EventGridEvent
			if r.Kind() == reflect.Ptr {
				v = r.Elem().Interface().(azfunc.EventGridEvent)
			} else {
				v = r.Interface().(azfunc.EventGridEvent)
			}

			data := map[string]interface{}{
				"requestId":   "71fd4516-701e-005b-0b38-135eb8000000",
				"contentType": "text/plain",
				"url":         "https://vladdbblobstorage.blob.core.windows.net/testcontainerblob/testblob.txt",
				"sequencer":   "00000000000000000000000000000F0A00000000000b67b2",
				"storageDiagnostics": map[string]interface{}{
					"batchId": "6f7a849b-7647-4e15-89de-962addd81215",
				},
				"api":             "PutBlockList",
				"eTag":            "0x8D5E14F9C71C823",
				"contentLength":   18.0,
				"blobType":        "BlockBlob",
				"clientRequestId": "58648a86-5e00-49fc-b1b1-e9bd6e98a025",
			}

			expected := azfunc.EventGridEvent{
				Data:            data,
				DataVersion:     "",
				EventTime:       time.Date(2018, 7, 4, 1, 43, 58, 617171500, time.UTC),
				EventType:       "Microsoft.Storage.BlobCreated",
				ID:              "71fd4516-701e-005b-0b38-135eb80633b3",
				MetadataVersion: "1",
				Subject:         "/blobServices/default/containers/testcontainerblob/blobs/testblob.txt",
				Topic:           "/subscriptions/7127e532-e730-40dd-acda-0ca1105c1e55/resourceGroups/valddFunctionGo/providers/Microsoft.Storage/storageAccounts/vladdbblobstorage",
			}

			got, want := v, expected
			if diff := deep.Equal(got, want); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func loadTestData(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func loadInvocationRequest(t *testing.T, name string) *rpc.InvocationRequest {
	b := loadTestData(t, name)
	r := bytes.NewReader(b)
	var ir rpc.InvocationRequest

	jsonpb.Unmarshal(r, &ir)
	return &ir
}
