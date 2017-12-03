package worker

import (
	log "github.com/Sirupsen/logrus"
	"github.com/radu-matei/azure-functions-golang-worker/executor"
	"github.com/radu-matei/azure-functions-golang-worker/rpc"
)

var (
	functionMap = make(map[string]*rpc.RpcFunctionMetadata)
)

func handleStreamingMessage(message *rpc.StreamingMessage, client *Client, eventStream rpc.FunctionRpc_EventStreamClient) {
	switch m := message.Content.(type) {

	case *rpc.StreamingMessage_WorkerInitRequest:
		handleWorkerInitRequest(message.RequestId, m, client, eventStream)

	case *rpc.StreamingMessage_FunctionLoadRequest:
		handleFunctionLoadRequest(message.RequestId, m, client, eventStream)

	case *rpc.StreamingMessage_InvocationRequest:
		handleInvocationRequest(message.RequestId, m, client, eventStream)

	default:
		log.Debugf("received message: %v", message)
	}
}

func handleWorkerInitRequest(requestID string,
	message *rpc.StreamingMessage_WorkerInitRequest,
	client *Client,
	eventStream rpc.FunctionRpc_EventStreamClient) {

	log.Debugf("received worker init request with host version %s ",
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
		log.Fatalf("Failed to send worker init response: %v", err)
	}
	log.Debugf("sent start worker init response : %v", workerInitResponse)
}

func handleFunctionLoadRequest(requestID string,
	message *rpc.StreamingMessage_FunctionLoadRequest,
	client *Client,
	eventStream rpc.FunctionRpc_EventStreamClient) {

	var status rpc.StatusResult_Status
	err := executor.LoadMethod(message.FunctionLoadRequest)
	if err != nil {
		status = rpc.StatusResult_Failure
	}

	status = rpc.StatusResult_Success

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
		log.Fatalf("Failed to send function load response: %v", err)
	}
	log.Debugf("sent function load response: %v", functionLoadResponse)
}

func handleInvocationRequest(requestID string,
	message *rpc.StreamingMessage_InvocationRequest,
	client *Client,
	eventStream rpc.FunctionRpc_EventStreamClient) {

	log.Debugf("received invocation request: %v", message.InvocationRequest)
	executor.ExecuteMethod(message.InvocationRequest)
}
