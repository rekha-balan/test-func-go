package cmd

import (
	"context"
	"fmt"
	"math"
	"os"

	"github.com/vladbarosan/test-func-go/internal/worker"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	flagDebug            bool
	host                 string
	port                 int
	workerID             string
	requestID            string
	grpcMaxMessageLength int
)

var rootCmd = &cobra.Command{
	Use:   "golangWorker",
	Short: "Runs the Azure Functions Golang Worker",
	Long: `The Azure Functions Golang Worker will initialize a connection with the Azure Functions runtime and
	will look for Golang plugins to load and it will dispatch calls to them.`,
	Run: func(cmd *cobra.Command, args []string) {
		startWorker(args)
	},
}

// Execute executes the given command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVar(&flagDebug, "debug", true, "enable verbose output")
	rootCmd.Flags().StringVar(&host, "host", "127.0.0.1", "RPC Server Host")
	rootCmd.Flags().IntVar(&port, "port", 0, "RPC Server Port")
	rootCmd.Flags().StringVar(&workerID, "workerId", "", "RPC Server Worker ID")
	rootCmd.Flags().StringVar(&requestID, "requestId", "", "Request ID")
	rootCmd.Flags().IntVar(&grpcMaxMessageLength, "grpcMaxMessageLength", math.MaxInt32, "Max message length")

	if flagDebug {
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(prefixFormatter{
			formatter: &log.TextFormatter{},
		})
	}
}

type prefixFormatter struct {
	formatter *log.TextFormatter
}

func (p prefixFormatter) Format(entry *log.Entry) ([]byte, error) {
	workerLoggerKeyPrefix := []byte("LanguageWorkerConsoleLog: ")
	f, err := p.formatter.Format(entry)
	if err != nil {
		return f, err
	}
	return append(workerLoggerKeyPrefix, f...), nil
}

func startWorker(args []string) {
	cfg := &worker.ClientConfig{
		Host:             host,
		Port:             port,
		WorkerID:         workerID,
		RequestID:        requestID,
		MaxMessageLength: grpcMaxMessageLength,
	}
	client := worker.NewClient(cfg)
	err := client.Connect()

	if err != nil {
		log.Fatalf("cannot create grpc connection: %v", err)
	}
	defer client.Disconnect()

	err = client.StartEventStream(context.Background())

	if err != nil {
		log.Fatalf("cannot start event stream: %v", err)
	}
}
