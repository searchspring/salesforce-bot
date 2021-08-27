package boost

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-storage-queue-go/azqueue"
	"github.com/kelseyhightower/envconfig"
	"github.com/searchspring/nebo/common"
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

// NewAzureStorage AzureStorage constructor that references ENV variables
func NewAzureStorage() AzureStorage {
	var env common.EnvVars
	envconfig.Process("", &env)

	return AzureStorage{
		AccountName:      env.AzureAccount,
		ConnectionString: env.AzureConnection,
	}
}

func (storage *AzureStorage) EnqueueMessage(queue string, cloudEvent CloudEvent) (string, error) {
	credentials, err := azqueue.NewSharedKeyCredential(storage.AccountName, storage.ConnectionString)

	if err != nil {
		log.Println("Failed to authenticate")
		return fmt.Sprintf("Failed to authenticate to Azure resource: `%v`", storage.AccountName), err
	}

	pipeline := azqueue.NewPipeline(credentials, azqueue.PipelineOptions{})

	u, _ := url.Parse(fmt.Sprintf("https://%s.queue.core.windows.net", storage.AccountName))
	serviceURL := azqueue.NewServiceURL(*u, pipeline)

	mainDispatchQueue := serviceURL.NewQueueURL(queue)
	ctx := context.TODO() // This example uses a never-expiring context.

	messagesURL := mainDispatchQueue.NewMessagesURL()

	bytes, _ := json.Marshal(cloudEvent)
	encodedStr := base64.StdEncoding.EncodeToString(bytes)

	var val string
	if _, err = messagesURL.Enqueue(ctx, encodedStr, time.Second*0, time.Minute); err == nil {
		val = "Success!"
	} else {
		val = "Failed"
		log.Println("Failure adding message to " + mainDispatchQueue.String())
	}
	return val, err
}
