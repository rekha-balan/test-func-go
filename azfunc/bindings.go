package azfunc

import (
	"context"
	"time"
)

// Context contains the runtime context of the function
type Context interface {
	context.Context
	FunctionID() string
	InvocationID() string
	Log(level int, format string, args ...interface{}) error
}

// LogLevel values
const (
	LogTrace = iota
	LogDebug
	LogInformation
	LogWarning
	LogError
	LogCritical
	LogNone
)

// Timer represents a timer trigger.
type Timer struct {
	PastDue       bool           `json:"IsPastDue"`
	ScheduleStats ScheduleStatus `json:"ScheduleStatus"`
}

// ScheduleStatus contains the schedule for a Timer.
type ScheduleStatus struct {
	Next        time.Time `json:"Next"`
	Last        time.Time `json:"Last"`
	LastUpdated time.Time `json:"LastUpdated"`
}

// QueueMsg represents an Azure queue message.
type QueueMsg struct {
	Text         string    `json:"azfuncdata"`
	ID           string    `json:"Id"`
	Insertion    time.Time `json:"InsertionTime"`
	Expiration   time.Time `json:"ExpirationTime"`
	PopReceipt   string    `json:"PopReceipt"`
	NextVisible  time.Time `json:"NextVisibleTime"`
	DequeueCount int       `json:"DequeueCount"`
}

// Blob contains the data from a blob as string.
type Blob struct {
	Content    string         `json:"azfuncdata"`
	Name       string         `json:"name"`
	URI        string         `json:"Uri"`
	Properties BlobProperties `json:"Properties"`
}

// BlobProperties contains metadata about a blob.
type BlobProperties struct {
	Length       int       `json:"Length"`
	ContentMD5   string    `json:"ContentMD5"`
	ContentType  string    `json:"ContentType"`
	ETag         string    `json:"ETag"`
	LastModified time.Time `json:"LastModified"`
}

//EventGridEvent represents properties of an event published to an Event Grid topic.
type EventGridEvent struct {
	// ID - An unique identifier for the event.
	ID string `json:"id"`
	// Topic - The resource path of the event source.
	Topic string `json:"topic"`
	// Subject - A resource path relative to the topic path.
	Subject string `json:"subject"`
	// Data - Event data specific to the event type.
	Data map[string]interface{} `json:"data"`
	// EventType - The type of the event that occurred.
	EventType string `json:"eventType"`
	// EventTime - The time (in UTC) the event was generated.
	EventTime time.Time `json:"eventTime"`
	// MetadataVersion - The schema version of the event metadata.
	MetadataVersion string `json:"metadataVersion"`
	// DataVersion - The schema version of the data object.
	DataVersion string `json:"dataVersion"`
}

// EventHubEvent represents properties of an event sent to an Event Hub.
type EventHubEvent struct {
	Data            string                 `json:"azfuncdata"`
	PartitionKey    *string                `json:"PartitionKey"`
	SequenceNumber  int                    `json:"SequenceNumber"`
	Offset          int                    `json:"Offset"`
	EnqueuedTimeUtc time.Time              `json:"EnqueuedTimeUtc"`
	Properties      map[string]interface{} `json:"Properties"`
}

// SBMsg represents a Service Bus Brokered Message.
type SBMsg struct {
	Data             string                 `json:"azfuncdata"`
	MessageID        string                 `json:"MessageId"`
	DeliveryCount    uint32                 `json:"DeliveryCount"`
	SequenceNumber   int64                  `json:"SequenceNumber"`
	ExpiresAtUtc     time.Time              `json:"ExpiresAtUtc"`
	EnqueuedTimeUtc  time.Time              `json:"EnqueuedTimeUtc"`
	ReplyTo          *string                `json:"ReplyTo"`
	To               *string                `json:"To"`
	CorrelationID    *string                `json:"CorrelationId"`
	Label            *string                `json:"Label"`
	ContentType      *string                `json:"ContentType"`
	DeadLetterSource *string                `json:"DeadLetterSource"`
	UserProperties   map[string]interface{} `json:"UserProperties"`
}
