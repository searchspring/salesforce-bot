package salesforce

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/nlopes/slack"
	"github.com/simpleforce/simpleforce"
	"searchspring.com/slack/validator"
)

// Platforms is a list of platforms in salesforce
var Platforms = []string{
	"3dcart",
	"BigCommerce",
	"CommerceV3",
	"Custom",
	"Magento",
	"Miva",
	"Netsuite",
	"Other",
	"Shopify",
	"Shopify Plus",
	"Yahoo",
}

type accountInfo struct {
	Website     string
	Manager     string
	Active      string
	MRR         float64
	FamilyMRR   float64
	Platform    string
	Integration string
	Provider    string
}

// DAO acts as the salesforce DAO
type DAO interface {
	Query(query string) ([]byte, error)
	IDQuery(query string) ([]byte, error)
	ResultToMessage(query string, result *simpleforce.QueryResult) ([]byte, error)
}

// DAOImpl defines the properties of the DAO
type DAOImpl struct {
	Client *simpleforce.Client
}

// NewDAO returns the salesforce DAO
func NewDAO(vars map[string]string) (DAO, error) {
	blanks := validator.FindBlankVals(vars)
	if len(blanks) > 0 {
		return nil, fmt.Errorf("the following env vars are not set: %s", strings.Join(blanks, ", "))
	}
	client := simpleforce.NewClient(vars["SF_URL"], simpleforce.DefaultClientID, simpleforce.DefaultAPIVersion)
	if client == nil {
		return nil, fmt.Errorf("nil returned from client creation")
	}
	err := client.LoginPassword(vars["SF_USER"], vars["SF_PASSWORD"], vars["SF_TOKEN"])
	if err != nil {
		return nil, err
	}
	return &DAOImpl{
		Client: client,
	}, nil
}

func (s *DAOImpl) Query(search string) ([]byte, error) {
	reg, err := regexp.Compile("[^a-zA-Z0-9_.-]+")
	if err != nil {
		return nil, err
	}

	sanitized := reg.ReplaceAllString(search, "")

	q := "SELECT Type, Website, CS_Manager__r.Name, Family_MRR__c, Chargify_MRR__c, Platform__c, Integration_Type__c, Chargify_Source__c " +
		"FROM Account WHERE Type IN ('Customer', 'Inactive Customer') " +
		"AND (Website LIKE '%" + sanitized + "%' OR Platform__c LIKE '%" + sanitized +
		"%' OR Tracking_Code__c = '" + sanitized + "') ORDER BY Chargify_MRR__c DESC"
	result, err := s.Client.Query(q)
	if err != nil {
		return nil, err
	}
	return s.ResultToMessage(sanitized, result)
}

func (s *DAOImpl) IDQuery(search string) ([]byte, error) {
	reg, err := regexp.Compile("[^a-zA-Z0-9_.-]+")
	if err != nil {
		return nil, err
	}

	sanitized := reg.ReplaceAllString(search, "")

	q := "SELECT Type, Website, CS_Manager__r.Name, Family_MRR__c, Chargify_MRR__c, Platform__c, Integration_Type__c, Chargify_Source__c " +
		"FROM Account WHERE Type IN ('Customer', 'Inactive Customer') AND Tracking_Code__c = '" + sanitized + "' ORDER BY Chargify_MRR__c DESC"
	result, err := s.Client.Query(q)
	if err != nil {
		return nil, err
	}
	return s.ResultToMessage(sanitized, result)
}

func (s *DAOImpl) ResultToMessage(search string, result *simpleforce.QueryResult) ([]byte, error) {
	accounts := []*accountInfo{}
	for _, record := range result.Records {
		manager := record["CS_Manager__r"]
		managerName := "unknown"
		if manager != nil {
			if mapName, ok := (manager.(map[string]interface{}))["Name"]; ok {
				managerName = fmt.Sprintf("%s", mapName)
			}
		}
		Type := record["Type"]
		active := "Active"
		if Type != "Customer" {
			active = "Not active"
		}
		platform := "unknown"
		if record["Platform__c"] != nil {
			platform = fmt.Sprintf("%s", record["Platform__c"])
		}
		integration := "unknown"
		if record["Integration_Type__c"] != nil {
			integration = fmt.Sprintf("%s", record["Integration_Type__c"])
		}
		provider := "unknown"
		if record["Chargify_Source__c"] != nil {
			provider = fmt.Sprintf("%s", record["Chargify_Source__c"])
		}
		mrr := float64(-1)
		if record["Chargify_MRR__c"] != nil {
			mrr = record["Chargify_MRR__c"].(float64)
		}
		familymrr := float64(-1)
		if record["Family_MRR__c"] != nil {
			familymrr = record["Family_MRR__c"].(float64)
		}

		accounts = append(accounts, &accountInfo{
			Website:     fmt.Sprintf("%s", record["Website"]),
			Manager:     fmt.Sprintf("%s", managerName),
			Active:      fmt.Sprintf("%s", active),
			MRR:         mrr,
			FamilyMRR:   familymrr,
			Platform:    platform,
			Integration: integration,
			Provider:    provider,
		})
	}
	accounts = cleanAccounts(accounts)
	if !isPlatformSearch(search) {
		accounts = sortAccounts(accounts)
	}
	accounts = truncateAccounts(accounts)
	msg := formatAccountInfos(accounts, search)
	return json.Marshal(msg)
}

func truncateAccounts(accounts []*accountInfo) []*accountInfo {
	truncated := []*accountInfo{}
	for i, account := range accounts {
		if i == 20 {
			break
		}
		truncated = append(truncated, account)
	}
	return truncated
}
func isPlatformSearch(search string) bool {
	for _, platform := range Platforms {
		if strings.ToLower(search) == strings.ToLower(platform) {
			return true
		}
	}
	return false
}

// example formatting here: https://api.slack.com/reference/messaging/attachments
func formatAccountInfos(accountInfos []*accountInfo, search string) *slack.Msg {
	initialText := "Reps for search: " + search
	if len(accountInfos) == 0 {
		initialText = "No results for: " + search
	}

	msg := &slack.Msg{
		ResponseType: slack.ResponseTypeInChannel,
		Text:         initialText,
		Attachments:  []slack.Attachment{},
	}
	for _, ai := range accountInfos {
		color := "3A23AD" // Searchspring purple
		if ai.Manager == "unknown" {
			color = "FF0000" // red
		}
		mrr := "unknown"
		if ai.MRR != -1 {
			mrr = fmt.Sprintf("$%.2f", ai.MRR)
		}
		familymrr := "unknown"
		if ai.FamilyMRR != -1 {
			familymrr = fmt.Sprintf("$%.2f", ai.FamilyMRR)
		}
		mrr = mrr + " (Family MRR: " + familymrr + ")"
		text := "Rep: " + ai.Manager + "\nMRR: " + mrr + "\nPlatform: " + ai.Platform + "\nIntegration: " + ai.Integration + "\nProvider: " + ai.Provider
		msg.Attachments = append(msg.Attachments, slack.Attachment{
			Color:      "#" + color,
			Text:       text,
			AuthorName: ai.Website + " (" + ai.Active + ")",
		})
	}
	return msg
}

func cleanAccounts(accounts []*accountInfo) []*accountInfo {
	for _, account := range accounts {
		w := account.Website
		if strings.HasPrefix(w, "http://") || strings.HasPrefix(w, "https://") {
			w = w[strings.Index(w, ":")+3:]
		}
		if strings.HasPrefix(w, "www.") {
			w = w[4:]
		}
		if strings.HasSuffix(w, "/") {
			w = w[0 : len(w)-1]
		}
		account.Website = w
	}
	return accounts
}
func sortAccounts(accounts []*accountInfo) []*accountInfo {
	sort.Slice(accounts, func(i, j int) bool {
		return len(accounts[i].Website) < len(accounts[j].Website)
	})
	return accounts
}
