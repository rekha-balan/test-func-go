package worker

import (
	"github.com/Azure/azure-functions-go/internal/rpc"
	"github.com/Azure/azure-functions-go/internal/runtime"
	log "github.com/sirupsen/logrus"
)

type worker struct {
	registry *runtime.Registry
}

// newWorker returns a new instance of Client
func newWorker() *worker {
	return &worker{
		registry: runtime.NewRegistry(),
	}
}

func (w worker) handleStreamingMessage(message *rpc.StreamingMessage, client *Client, eventStream rpc.FunctionRpc_EventStreamClient) {
	log.Debugf("received message: %v", message)
	switch m := message.Content.(type) {

	case *rpc.StreamingMessage_WorkerInitRequest:
		w.handleWorkerInitRequest(message.RequestId, m, client, eventStream)

	case *rpc.StreamingMessage_FunctionLoadRequest:
		w.handleFunctionLoadRequest(message.RequestId, m, client, eventStream)

	case *rpc.StreamingMessage_InvocationRequest:
		w.handleInvocationRequest(message.RequestId, m, client, eventStream)

	default:
		log.Debugf("received message: %v", message)
	}
}

func (w worker) handleWorkerInitRequest(requestID string,
	message *rpc.StreamingMessage_WorkerInitRequest,
	client *Client,
	eventStream rpc.FunctionRpc_EventStreamClient) {

	log.Debugf("received worker init request with host version %s",
		message.WorkerInitRequest.HostVersion)

	workerInitResponse := &rpc.StreamingMessage{
		RequestId: requestID,
		Content: &rpc.StreamingMessage_WorkerInitResponse{
			WorkerInitResponse: &rpc.WorkerInitResponse{
				Result: &rpc.StatusResult{
					Status: rpc.StatusResult_Success,
				},
			},
		},
	}

	if err := eventStream.Send(workerInitResponse); err != nil {
		log.Fatalf("failed to send worker init response: %v", err)
	}
	log.Debugf("sent start worker init response: %v", workerInitResponse)
}

func (w worker) handleFunctionLoadRequest(requestID string,
	message *rpc.StreamingMessage_FunctionLoadRequest,
	client *Client,
	eventStream rpc.FunctionRpc_EventStreamClient) {

	status := rpc.StatusResult_Success
	err := w.registry.LoadFunc(message.FunctionLoadRequest)
	if err != nil {
		status = rpc.StatusResult_Failure
		log.Debugf("could not load function: %v", err)
	}

	functionLoadResponse := &rpc.StreamingMessage{
		RequestId: requestID,
		Content: &rpc.StreamingMessage_FunctionLoadResponse{
			FunctionLoadResponse: &rpc.FunctionLoadResponse{
				FunctionId: message.FunctionLoadRequest.FunctionId,
				Result: &rpc.StatusResult{
					Status: status,
				},
			},
		},
	}

	if err := eventStream.Send(functionLoadResponse); err != nil {
		log.Fatalf("failed to send function load response: %v", err)
	}
	log.Debugf("sent function load response: %v", functionLoadResponse)
}

func (w worker) handleInvocationRequest(requestID string,
	message *rpc.StreamingMessage_InvocationRequest,
	client *Client,
	eventStream rpc.FunctionRpc_EventStreamClient) {

	response := w.registry.ExecuteFunc(message.InvocationRequest, eventStream)

	invocationResponse := &rpc.StreamingMessage{
		RequestId: requestID,
		Content: &rpc.StreamingMessage_InvocationResponse{
			InvocationResponse: response,
		},
	}

	if err := eventStream.Send(invocationResponse); err != nil {
		log.Fatalf("failed to send function invocation response: %v", err)
	}
}
