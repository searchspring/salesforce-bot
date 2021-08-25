package models

// AzureQueueMessage The queue message format that gets sent to the Boost teams "dispatch-main" queue
// which then re-distributes the message to any workers queues
type AzureQueueMessage struct {
	Id          string      `json:id`
	Type        string      `json:type`
	Source      string      `json:source`
	Subject     string      `json:subject`
	Specversion string      `json:specversion`
	Time        string      `json:time`
	Data        interface{} `json:data,omitempty`
}