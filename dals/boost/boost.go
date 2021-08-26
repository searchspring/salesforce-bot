package boost

import (
	"encoding/json"
	"fmt"
	"github.com/searchspring/nebo/common"
	"log"
	"net/http"
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
	return GenericJsonRequest(client, url)
}

func HandleGetExclusionStatsRequest(trackingCode string, client *common.Client) map[string]interface{} {
	url := fmt.Sprintf("%v/sites/%v/exclusionStats", boostAdminUrl, trackingCode)
	return GenericJsonRequest(client, url)
}

// GenericJsonRequest Make a request to a URL and return the *flat* list of JSON back to the caller as a map
func GenericJsonRequest(client *common.Client, url string) map[string]interface{} {
	var values = make(map[string]interface{})

	if bytes, err := client.Get(url); err == nil {
		json.Unmarshal(bytes, &values)
	} else {
		log.Println("Error parsing response from " + url)
	}
	return values
}

func RestartSite(trackingCode string) {
	url := fmt.Sprintf("%v/sites/%v/restart", boostAdminUrl, trackingCode)
	http.Post(url, "application/json", nil)
}

func HandlePauseRequest(trackingCode string) (string, error) {
	azure := NewAzureStorage()
	cloudEvent := NewCloudEvent("searchspring.boost.pauseRequested", trackingCode, nil)
	return azure.EnqueueMessage("admin", cloudEvent)
}

func HelpText() string {
	return "Try a command from this here list:\n\n" +
		strings.Join(SlackOptions, "\n")
}
