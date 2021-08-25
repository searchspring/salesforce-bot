package boost

import (
	"github.com/google/uuid"
	"time"
)

// CloudEvent The queue message format that gets sent to the Boost teams "dispatch-main" queue
// `dispatch-main` then re-distributes the message to one, or many, workers queues
type CloudEvent struct {
	Id      string      `json:"id"`
	Type    string      `json:"type"`
	Source  string      `json:"source"`
	Subject string      `json:"subject"`
	Version string      `json:"specversion"`
	Time    string      `json:"time"`
	Data    interface{} `json:"data,omitempty":`
}

// NewCloudEvent CloudEvent constructor with some default values
func NewCloudEvent(eventType string, subject string, data interface{}) CloudEvent {
	return CloudEvent{
		Id:      uuid.New().String(),
		Source:  "/nebo",
		Time:    utcTimeNow(),
		Version: "1.0",
		Type:    eventType,
		Subject: subject,
		Data:    data,
	}
}

func utcTimeNow() string {
	return time.Now().UTC().Format(time.RFC3339)
}
