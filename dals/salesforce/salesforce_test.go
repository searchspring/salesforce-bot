package salesforce

import (
	"encoding/json"
	"testing"

	"github.com/searchspring/nebo/common"
	"github.com/searchspring/nebo/models"
	"github.com/simpleforce/simpleforce"
	"github.com/stretchr/testify/require"
)

func createQueryResults() *simpleforce.QueryResult {
	qr := &simpleforce.QueryResult{}
	json.Unmarshal([]byte(`{ "totalSize": 1,
        "done": true,
        "records": [{ 
                "Website": "fabletics.com",
                "CS_Manager__r": { "Name": "Ashley Hilton" },
                "Family_MRR__c": 14858.54,
                "Chargify_MRR__c": 3955.17,
                "Integration_Type__c":"v3",
                "Chargify_Source__c":"Searchspring",
                "Platform__c":"Custom",
                "Tracking_Code__c": "wub9gl",
                "BillingCity": "Chicago",
                "BillingState": "IL"} 
            ]
        }`), qr)
	return qr
}

func TestResultToMessage(t *testing.T) {
	dao := &DAOImpl{}
	response, err := dao.ResultToMessage("search term", createQueryResults())
	require.Nil(t, err)
	require.Contains(t, response[0].Manager, "Ashley Hilton")
	require.Contains(t, response[0].Website, "fabletics.com")
	require.Contains(t, response[0].Active, "Not active")
	require.Equal(t, response[0].MRR, 3955.17)
	require.Equal(t, response[0].FamilyMRR, 14858.54)
	require.Contains(t, response[0].Platform, "Custom")
	require.Contains(t, response[0].Integration, "v3")
	require.Contains(t, response[0].Provider, "Searchspring")
	require.Contains(t, response[0].SiteId, "wub9gl")
	require.Contains(t, response[0].City, "Chicago")
	require.Contains(t, response[0].State, "IL")
}

func TestFormatAccountInfos(t *testing.T) {
	var response []*models.AccountInfo

	account := &models.AccountInfo{
		Website:     "fabletics.com",
		Manager:     "Ashley Hilton",
		Active:      "Not active",
		MRR:         3955.17,
		FamilyMRR:   14858.54,
		Platform:    "Custom",
		Integration: "v3",
		Provider:    "Searchspring",
		SiteId:      "wub9gl",
		City:        "Chicago",
		State:       "IL",
	}
	response = append(response, account)
	msg := common.FormatAccountInfos(response, "fabletics")
	require.Contains(t, msg.Attachments[0].Text, "Rep: Ashley Hilton")
	require.Contains(t, msg.Attachments[0].Text, "MRR: $3,955.17")
	require.Contains(t, msg.Attachments[0].Text, "Platform: Custom")
	require.Contains(t, msg.Attachments[0].Text, "Integration: v3")
	require.Contains(t, msg.Attachments[0].Text, "Provider: Searchspring")
	require.Contains(t, msg.Attachments[0].Text, "Family MRR: $14,858.54")
	require.Contains(t, msg.Attachments[0].Text, "Location: Chicago, IL")
	require.Equal(t, "fabletics.com (Not active) (SiteId: wub9gl)", msg.Attachments[0].AuthorName)
	require.Equal(t, "#3A23AD", msg.Attachments[0].Color)
}
