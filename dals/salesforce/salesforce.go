package salesforce

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	common "github.com/searchspring/nebo/common"
	"github.com/simpleforce/simpleforce"
)

// DAO acts as the salesforce DAO
type DAO interface {
	Query(query string) ([]*common.AccountInfo, error)
	IDQuery(search string) ([]*common.AccountInfo, error)
	ResultToMessage(query string, result *simpleforce.QueryResult) ([]*common.AccountInfo, error)
	NPSQuery(query string) ([]*common.AccountInfo, error)
	StructFromResult(query string, result *simpleforce.QueryResult) ([]*common.AccountInfo, error)
	GetSearchKey() string
}

// DAOImpl defines the properties of the DAO
type DAOImpl struct {
	Client *simpleforce.Client
}

const selectFields = "Type, Website, CS_Manager__r.Name, Family_MRR__c, Chargify_MRR__c, Platform__c, Integration_Type__c, Chargify_Source__c, Tracking_Code__c, BillingCity, BillingCountry, BillingState"

// NewDAO returns the salesforce DAO
func NewDAO(sfURL string, sfUser string, sfPassword string, sfToken string) DAO {
	if common.ContainsEmptyString(sfURL, sfUser, sfPassword, sfToken) {
		return nil
	}
	client := simpleforce.NewClient(sfURL, simpleforce.DefaultClientID, simpleforce.DefaultAPIVersion)
	if client == nil {
		log.Println("nil returned from client creation")
		return nil
	}
	err := client.LoginPassword(sfUser, sfPassword, sfToken)
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	return &DAOImpl{
		Client: client,
	}
}

func (s *DAOImpl) Query(search string) ([]*common.AccountInfo, error) {
	reg, err := regexp.Compile("[^a-zA-Z0-9_.-]+")
	if err != nil {
		return nil, err
	}

	sanitized := reg.ReplaceAllString(search, "")

	q := "SELECT " + selectFields + " " +
		"FROM Account WHERE Type IN ('Customer', 'Inactive Customer') " +
		"AND (Website LIKE '%" + sanitized + "%' OR Platform__c LIKE '%" + sanitized +
		"%' OR Tracking_Code__c = '" + sanitized + "') ORDER BY Chargify_MRR__c DESC"
	result, err := s.Client.Query(q)
	if err != nil {
		return nil, err
	}
	return s.ResultToMessage(sanitized, result)
}

func (s *DAOImpl) IDQuery(search string) ([]*common.AccountInfo, error) {
	reg, err := regexp.Compile("[^a-zA-Z0-9_.-]+")
	if err != nil {
		return nil, err
	}

	sanitized := reg.ReplaceAllString(search, "")

	q := "SELECT " + selectFields + " " +
		"FROM Account WHERE Type IN ('Customer', 'Inactive Customer') AND Tracking_Code__c = '" + sanitized + "' ORDER BY Chargify_MRR__c DESC"
	result, err := s.Client.Query(q)
	if err != nil {
		return nil, err
	}
	return s.ResultToMessage(sanitized, result)
}

func (s *DAOImpl) ResultToMessage(search string, result *simpleforce.QueryResult) ([]*common.AccountInfo, error) {
	accounts := []*common.AccountInfo{}
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
		siteId := "unknown"
		if record["Tracking_Code__c"] != nil {
			siteId = fmt.Sprintf("%s", record["Tracking_Code__c"])
		}
		city := "unknown"
		state := "unknown"
		if record["BillingCity"] != nil && record["BillingState"] != nil {
			city = fmt.Sprintf("%s", record["BillingCity"])
			state = fmt.Sprintf("%s", record["BillingState"])
		}

		accounts = append(accounts, &common.AccountInfo{
			Website:     fmt.Sprintf("%s", record["Website"]),
			Manager:     managerName,
			Active:      active,
			MRR:         mrr,
			FamilyMRR:   familymrr,
			Platform:    platform,
			Integration: integration,
			Provider:    provider,
			SiteId:      siteId,
			City:        city,
			State:       state,
		})
	}
	accounts = cleanAccounts(accounts)
	if !isPlatformSearch(search) {
		accounts = sortAccounts(accounts)
	}
	accounts = truncateAccounts(accounts)
	return accounts, nil
}

// nps functions
func (s *DAOImpl) NPSQuery(search string) ([]*common.AccountInfo, error) {
	reg, err := regexp.Compile("[^a-zA-Z0-9_.-]+")
	if err != nil {
		return nil, err
	}

	sanitized := reg.ReplaceAllString(search, "")

	q := "SELECT " + selectFields + " " +
		"FROM Account WHERE Type IN ('Customer', 'Inactive Customer') " +
		"AND (Website LIKE '%" + sanitized + "%' OR Platform__c LIKE '%" + sanitized +
		"%' OR Tracking_Code__c = '" + sanitized + "') ORDER BY Chargify_MRR__c DESC"
	result, err := s.Client.Query(q)
	if err != nil {
		return nil, err
	}
	return s.StructFromResult(sanitized, result)
}

func (s *DAOImpl) StructFromResult(search string, result *simpleforce.QueryResult) ([]*common.AccountInfo, error) {
	accounts := []*common.AccountInfo{}
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
		mrr := float64(-1)
		if record["Chargify_MRR__c"] != nil {
			mrr = record["Chargify_MRR__c"].(float64)
		}

		familymrr := float64(-1)
		if record["Family_MRR__c"] != nil {
			familymrr = record["Family_MRR__c"].(float64)
		}

		accounts = append(accounts, &common.AccountInfo{
			Manager:   managerName,
			Active:    active,
			MRR:       mrr,
			FamilyMRR: familymrr,
		})
	}
	accounts = cleanAccounts(accounts)
	if !isPlatformSearch(search) {
		accounts = sortAccounts(accounts)
	}
	accounts = truncateAccounts(accounts)
	return accounts, nil
}

func (s *DAOImpl) GetSearchKey() string {
	return ""
}

func truncateAccounts(accounts []*common.AccountInfo) []*common.AccountInfo {
	truncated := []*common.AccountInfo{}
	for i, account := range accounts {
		if i == 20 {
			break
		}
		truncated = append(truncated, account)
	}
	return truncated
}
func isPlatformSearch(search string) bool {
	for _, platform := range common.Platforms {
		if strings.EqualFold(search, platform) {
			return true
		}
	}
	return false
}

func cleanAccounts(accounts []*common.AccountInfo) []*common.AccountInfo {
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
func sortAccounts(accounts []*common.AccountInfo) []*common.AccountInfo {
	sort.Slice(accounts, func(i, j int) bool {
		return len(accounts[i].Website) < len(accounts[j].Website)
	})
	return accounts
}
