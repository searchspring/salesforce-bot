package boost

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Site BoostAdminApi response from /sites?status=hung
type Site struct {
	Status, Message, SiteId, Name string
}

const boostAdminUrl = "https://boostadmin.azurewebsites.net"

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

func HandleStatusRequest(trackingCode string) map[string]interface{} {
	url := boostAdminUrl + "/sites/" + trackingCode + "/status"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error getting status for " + trackingCode)
	}
	defer resp.Body.Close()
	body, err2 := io.ReadAll(resp.Body)
	var status = make(map[string]interface{})

	if err2 != nil {
		fmt.Println("Error parsing get status response")
	} else {
		json.Unmarshal(body, &status)
	}
	return status
}

func HandleGetExclusionStatsRequest(trackingCode string) map[string]interface{} {
	url := boostAdminUrl + "/sites/" + trackingCode + "/exclusionStats"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error getting exclusionStats for " + trackingCode)
	}
	defer resp.Body.Close()
	body, err2 := io.ReadAll(resp.Body)
	var stats = make(map[string]interface{})

	if err2 != nil {
		fmt.Println("Error parsing exclusion stats response")
	} else {
		json.Unmarshal(body, &stats)
	}
	return stats
}

func RestartSite(trackingCode string) {
	url := boostAdminUrl + "/sites/" + trackingCode + "/restart"
	_, err := http.Post(url, "application/json", nil)
	if err != nil {
		fmt.Println("Error restarting " + trackingCode)
	}
	fmt.Println("Restarted " + trackingCode)
}

func getHungSites() []Site {
	resp, err := http.Get(boostAdminUrl + "/sites?status=hung")
	if err != nil {
		fmt.Println("Error getting list of hung sites")
	}

	defer resp.Body.Close()
	body, err2 := io.ReadAll(resp.Body)

	if err2 != nil {
		fmt.Println("Error parsing list of sites")
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
