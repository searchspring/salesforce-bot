package boost

import (
	"encoding/json"
	"fmt"
	"github.com/searchspring/nebo/common"
	"io"
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

func RestartHungSites() {
	restartSites(getHungSites())
}

func restartSites(sites []Site) {
	if sites != nil {
		for _, site := range sites {
			if len(site.SiteId) == 6 {
				RestartSite(site.SiteId)
			}
		}
	}
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
	_, err := http.Post(url, "application/json", nil)
	if err != nil {
		log.Println("Error restarting " + trackingCode)
	}
	log.Println("Restarted " + trackingCode)
}

func HandlePauseRequest(trackingCode string) (string, error) {
	azure := NewAzureStorage()
	cloudEvent := NewCloudEvent("searchspring.boost.pauseRequested", trackingCode, nil)
	return azure.EnqueueMessage("testing", cloudEvent)
}

func getHungSites() []Site {
	resp, err := http.Get(boostAdminUrl + "/sites?status=hung")
	if err != nil {
		log.Println("Error getting list of hung sites")
	}

	defer resp.Body.Close()
	body, err2 := io.ReadAll(resp.Body)

	if err2 != nil {
		log.Println("Error parsing list of sites")
	} else {
		var sites []Site
		json.Unmarshal(body, &sites)
		return sites
	}
	return nil
}

func HelpText() string {
	return "Try a command from this here list:\n\n" +
		strings.Join(SlackOptions, "\n")
}
