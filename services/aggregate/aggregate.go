package aggregate

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/searchspring/nebo/common"
	"github.com/searchspring/nebo/dals/metabase"
	"github.com/searchspring/nebo/dals/salesforce"
	"github.com/searchspring/nebo/models"
)

type Deps struct {
	MetabaseDAO   metabase.DAO
	SalesforceDAO salesforce.DAO
}

type AggregateService interface {
	Query(query string) ([]byte, error)
}

type AggregateServiceImpl struct {
	Deps *Deps
}

func (d *AggregateServiceImpl) Query(search string) ([]byte, error) {
	metabaseData, err := d.Deps.MetabaseDAO.Query(search)
	if err != nil {
		return nil, nil
	}
	salesforceData, err := d.Deps.SalesforceDAO.Query(search)
	if err != nil {
		return nil, nil
	}

	aggregatedData := addMetabaseAccounts(metabaseData, salesforceData)
	aggregatedData = addSalesforceAccounts(aggregatedData, salesforceData)

	aggregatedData = cleanAccounts(aggregatedData)
	if !isPlatformSearch(search) {
		aggregatedData = sortAccounts(aggregatedData, "website")
	}
	aggregatedData = truncateToTwenty(aggregatedData)
	aggregatedData = sortAccounts(aggregatedData, "mrr")

	msg := common.FormatAccountInfos(aggregatedData, search)
	return json.Marshal(msg)
}

func (d *AggregateServiceImpl) QueryPartners(search string) ([]byte, error) {
	salesforceData, err := d.Deps.SalesforceDAO.QueryPartners(search)
	if err != nil {
		return nil, nil
	}
	msg := common.FormatPartnerInfos(salesforceData, search)
	return json.Marshal(msg)
}

// helper functions

func addMetabaseAccounts(metabaseData []*models.AccountInfo, salesforceData []*models.AccountInfo) []*models.AccountInfo {
	var customerData []*models.AccountInfo
	for _, v := range metabaseData {
		e, i := exists(v.SiteId, v.Website, salesforceData)
		if e {
			if salesforceData[i].Type == "Customer" || salesforceData[i].Type == "Inactive Customer" {
				customerData = append(customerData, v)
			}
		} else {
			customerData = append(customerData, v)
		}
	}
	return customerData
}

func addSalesforceAccounts(currentCustomerData []*models.AccountInfo, salesforceData []*models.AccountInfo) []*models.AccountInfo {
	for _, v := range salesforceData {
		e, _ := exists(v.SiteId, v.Website, currentCustomerData)
		if !e {
			if v.Type == "Customer" || v.Type == "Inactive Customer" {
				currentCustomerData = append(currentCustomerData, v)
			}
		}
	}
	return currentCustomerData
}

func exists(id string, website string, data []*models.AccountInfo) (result bool, index int) {
	for i, account := range data {
		if account.Website == website || account.SiteId == id && id != "unknown" {
			return true, i
		}
	}
	return false, -1
}

// cleaning account arrays

func truncateToTwenty(accounts []*models.AccountInfo) []*models.AccountInfo {
	truncated := []*models.AccountInfo{}
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

func cleanAccounts(accounts []*models.AccountInfo) []*models.AccountInfo {
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

func sortAccounts(accounts []*models.AccountInfo, sortType string) []*models.AccountInfo {
	sort.Slice(accounts, func(i, j int) bool {
		if sortType == "website" {
			return len(accounts[i].Website) < len(accounts[j].Website)
		} else {
			return accounts[i].MRR > accounts[j].MRR
		}
	})
	return accounts
}
