package salesforce

import (
	"fmt"
	"log"
	"regexp"

	common "github.com/searchspring/nebo/common"
	"github.com/searchspring/nebo/models"
	"github.com/simpleforce/simpleforce"
)

// DAO acts as the salesforce DAO
type DAO interface {
	Query(query string) ([]*models.AccountInfo, error)
	QueryPartners(query string) ([]*models.PartnerInfo, error)
	ResultToMessage(query string, result *simpleforce.QueryResult) ([]*models.AccountInfo, error)
	ResultToPartnerMessage(search string, result *simpleforce.QueryResult) ([]*models.PartnerInfo, error)
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

func (s *DAOImpl) Query(search string) ([]*models.AccountInfo, error) {
	sanitized, err := sanitize(search)
	if err != nil {
		return nil, err
	}
	q := "SELECT " + selectFields + " " +
		"FROM Account WHERE (Website LIKE '%" + sanitized + "%' OR Platform__c LIKE '%" + sanitized +
		"%' OR Tracking_Code__c = '" + sanitized + "') ORDER BY Chargify_MRR__c DESC"
	result, err := s.Client.Query(q)
	if err != nil {
		return nil, err
	}
	return s.ResultToMessage(sanitized, result)
}

func sanitize(search string) (sanitizedSearch string, err error) {
	reg, err := regexp.Compile("[^a-zA-Z0-9_.-]+")
	if err != nil {
		return sanitizedSearch, err
	}
	return reg.ReplaceAllString(search, ""), nil
}

func (s *DAOImpl) QueryPartners(search string) ([]*models.PartnerInfo, error) {
	sanitized, err := sanitize(search)
	if err != nil {
		return nil, err
	}
	q := "SELECT Name, Type, Account_Status__c, OwnerId, Partner_Type__c, Supported_Platforms__c, Partner_Terms__c, Partner_Terms_Notes__c " +
		"FROM Account WHERE (Type = 'Partner' OR Type = 'Potential Partner') AND Name LIKE '%" + sanitized + "%' " +
		"ORDER BY Name LIMIT 5"
	result, err := s.Client.Query(q)
	if err != nil {
		return nil, err
	}
	return s.ResultToPartnerMessage(sanitized, result)
}

func getField(field interface{}) (result string) {
	result = "unknown"
	if field != nil {
		result = fmt.Sprintf("%s", field)
	}
	return result
}

func (s *DAOImpl) ResultToPartnerMessage(search string, result *simpleforce.QueryResult) ([]*models.PartnerInfo, error) {
	partners := []*models.PartnerInfo{}
	for _, record := range result.Records {
		partners = append(partners, &models.PartnerInfo{
			Name:               getField(record["Name"]),
			Type:               getField(record["Type"]),
			Status:             getField(record["Account_Status__c"]),
			OwnerID:            getField(record["OwnerId"]),
			PartnerType:        getField(record["Partner_Type__c"]),
			SupportedPlatforms: getField(record["Supported_Platforms__c"]),
			PartnerTerms:       getField(record["Partner_Terms__c"]),
			PartnerTermsNotes:  getField(record["Partner_Terms_Notes__c"]),
		})
	}
	return partners, nil
}

func (s *DAOImpl) ResultToMessage(search string, result *simpleforce.QueryResult) ([]*models.AccountInfo, error) {
	accounts := []*models.AccountInfo{}
	for _, record := range result.Records {
		manager := record["CS_Manager__r"]
		managerName := "unknown"
		if manager != nil {
			if mapName, ok := (manager.(map[string]interface{}))["Name"]; ok {
				managerName = fmt.Sprintf("%s", mapName)
			}
		}
		Type := fmt.Sprintf("%s", record["Type"])
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

		accounts = append(accounts, &models.AccountInfo{
			Website:     fmt.Sprintf("%s", record["Website"]),
			Manager:     managerName,
			Active:      active,
			Type:        Type,
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

	return accounts, nil
}

func (s *DAOImpl) GetSearchKey() string {
	return ""
}
