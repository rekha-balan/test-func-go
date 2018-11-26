package worker

import (
	"context"
	"fmt"
	"io"

	"github.com/vladbarosan/test-func-go/internal/rpc"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// ClientConfig contains all necessary configuration to connect to the Azure Functions Host
type ClientConfig struct {
	Host             string
	Port             int
	WorkerID         string
	RequestID        string
	MaxMessageLength int
}

// Client that listens for events from the Azure Functions host and executes Golang methods
type Client struct {
	Cfg    *ClientConfig
	conn   *grpc.ClientConn
	worker *worker
}

// NewClient returns a new instance of Client
func NewClient(cfg *ClientConfig) *Client {
	return &Client{
		Cfg:    cfg,
		worker: newWorker(),
	}
}

// StartEventStream starts listening for messages from the Azure Functions Host
func (c *Client) StartEventStream(ctx context.Context, opts ...grpc.CallOption) error {
	log.Debugf("starting event stream..")
	eventStream, err := rpc.NewFunctionRpcClient(c.conn).EventStream(ctx)
	if err != nil {
		log.Fatalf("cannot get event stream: %v", err)
		return err
	}

	waitc := make(chan struct{})
	go func() {
		for {
			message, err := eventStream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("error receiving stream: %v", err)
				continue
			}

			go c.worker.handleStreamingMessage(message, c, eventStream)
		}
	}()

	startStreamingMessage := &rpc.StreamingMessage{
		RequestId: c.Cfg.RequestID,
		Content: &rpc.StreamingMessage_StartStream{
			StartStream: &rpc.StartStream{
				WorkerId: c.Cfg.WorkerID,
			},
		},
	}

	if err = eventStream.Send(startStreamingMessage); err != nil {
		log.Fatalf("failed to send start streaming request: %v", err)
		return err
	}
	log.Debugf("sent start streaming message to host")

	<-waitc
	return nil
}

// Connect tries to establish a grpc connection with the server
func (c *Client) Connect(opts ...grpc.DialOption) (err error) {
	log.Debugf("attempting to start grpc connection to server %s:%d with worker id %s and request id %s", c.Cfg.Host, c.Cfg.Port, c.Cfg.WorkerID, c.Cfg.RequestID)
	opts = append(opts, grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(c.Cfg.MaxMessageLength)))

	conn, err := c.getGRPCConnection(opts)
	if err != nil {
		log.Fatalf("cannot create grpc connection: %v", err)
		return
	}

	c.conn = conn
	log.Debugf("started grpc connection...")
	return
}

// Disconnect closes the connection to the server
func (c *Client) Disconnect() error {
	return c.conn.Close()
}

//getGRPCConnection returns a new grpc connection
func (c *Client) getGRPCConnection(opts []grpc.DialOption) (conn *grpc.ClientConn, err error) {
	host := fmt.Sprintf("%s:%d", c.Cfg.Host, c.Cfg.Port)
	log.Debugf("trying to dial %s", host)
	if conn, err = grpc.Dial(host, opts...); err != nil {
		return nil, fmt.Errorf("failed to dial %q: %v", host, err)
	}
	return conn, nil
}
