package boost

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-storage-queue-go/azqueue"
	"github.com/searchspring/nebo/models"
	"log"
	"net/url"
	"time"
)

// Golang Azure Queue docs
// https://pkg.go.dev/github.com/Azure/azure-storage-queue-go/azqueue?utm_source=godoc#pkg-overview

type AzureStorage struct {
	AccountName      string
	ConnectionString string
}

func (storage *AzureStorage) EnqueueMessage(queue string, action string, subject string) (string, error) {
	credentials, err := azqueue.NewSharedKeyCredential(storage.AccountName, storage.ConnectionString)

	if err != nil {
		log.Println("Failed to authenticate")
		return fmt.Sprintf("Failed to authenticate to Azure resource: `%v`", storage.AccountName), err
	}
	log.Println("Authenticated to Azure..")

	pipeline := azqueue.NewPipeline(credentials, azqueue.PipelineOptions{})

	u, _ := url.Parse(fmt.Sprintf("https://%s.queue.core.windows.net", storage.AccountName))
	serviceURL := azqueue.NewServiceURL(*u, pipeline)

	mainDispatchQueue := serviceURL.NewQueueURL(queue)
	ctx := context.TODO() // This example uses a never-expiring context.

	messagesURL := mainDispatchQueue.NewMessagesURL()
	payload := models.AzureQueueMessage{
		Source:      "/nebo",
		Type:        action,
		Subject:     subject,
		Data:        nil,
	}

	b, _ := json.Marshal(payload)
	_, err = messagesURL.Enqueue(ctx, string(b), time.Second*0, time.Minute)
	var val string
	if err == nil {
		val = "Success!"
	} else {
		val = "Failed"
		log.Println("Failure adding message to " + mainDispatchQueue.String())
	}
	return val, err
}
