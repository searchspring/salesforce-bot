package boost

import (
	"encoding/json"
	"fmt"
	"github.com/searchspring/nebo/common"
	"log"
	"strings"
)

// Site BoostAdminApi response from /sites?status=hung
type Site struct {
	Status, Message, SiteId, Name string
}

type UpdateResponse struct {
	Status string `json:"status"`
}

func HandleGetStatusRequest(trackingCode string, client *common.Client) map[string]interface{} {
	url := fmt.Sprintf("%v/sites/%v/status", boostAdminUrl, trackingCode)
	return genericJsonRequest(client, url)
}

func HandleGetExclusionStatsRequest(trackingCode string, client *common.Client) map[string]interface{} {
	url := fmt.Sprintf("%v/sites/%v/exclusionStats", boostAdminUrl, trackingCode)
	return genericJsonRequest(client, url)
}

// GenericJsonRequest Make a request to a URL and return the *flat* list of JSON back to the caller as a map
func genericJsonRequest(client *common.Client, url string) map[string]interface{} {
	var values = make(map[string]interface{})

	if bytes, err := client.Get(url); err == nil {
		json.Unmarshal(bytes, &values)
	} else {
		log.Println("Error parsing response from " + url)
	}
	return values
}

func HandleUpdateRequest(trackingCode string) (string, error) {
	azure := NewAzureStorage()
	cloudEvent := NewCloudEvent("searchspring.boost.updateRequested", trackingCode, nil)
	return azure.EnqueueMessage(mainBoostDispatchQueue, cloudEvent)
}

func HandlePauseUpdates(trackingCode string) (string, error) {
	azure := NewAzureStorage()
	cloudEvent := NewCloudEvent("searchspring.boost.pauseUpdatesRequested", trackingCode, nil)
	return azure.EnqueueMessage(mainBoostDispatchQueue, cloudEvent)
}

func HandleResumeUpdates(trackingCode string) (string, error) {
	azure := NewAzureStorage()
	cloudEvent := NewCloudEvent("searchspring.boost.resumeUpdatesRequested", trackingCode, nil)
	return azure.EnqueueMessage(mainBoostDispatchQueue, cloudEvent)
}

func HandleCancelRequest(trackingCode string) (string, error) {
	azure := NewAzureStorage()
	cloudEvent := NewCloudEvent("searchspring.boost.cancelRequested", trackingCode, nil)
	return azure.EnqueueMessage(mainBoostDispatchQueue, cloudEvent)
}

func HelpText() string {
	return "Try a command from this here list:\n\n" +
		strings.Join(SlackOptions, "\n")
}
