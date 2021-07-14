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
	var aggregatedData []*models.AccountInfo

	metabaseData, err := d.Deps.MetabaseDAO.Query(search)
	if err != nil {
		return nil, nil
	}
	salesforceData, err := d.Deps.SalesforceDAO.Query(search)
	if err != nil {
		return nil, nil
	}

	/* add all metabase data to the final array that A) Doesn't have a corrisponding document in SF
	   B) Has a corrisponding document in SF as long as the SF document is of type Customer or Inactive Customer
	*/
	for _, v := range metabaseData {
		e, i := exists(v.SiteId, v.Website, salesforceData)
		if e {
			if salesforceData[i].Type == "Customer" || salesforceData[i].Type == "Inactive Customer" {
				aggregatedData = append(aggregatedData, v)
			}
		} else {
			aggregatedData = append(aggregatedData, v)
		}
	}

	// add all salesforce data to the final array that doesn't already exist in the final array and is of type Customer or Inactive Customer
	for _, v := range salesforceData {
		e, _ := exists(v.SiteId, v.Website, aggregatedData) 
		if !e {
			if v.Type == "Customer" || v.Type == "Inactive Customer" {
				aggregatedData = append(aggregatedData, v)
			}
		}
	}

	aggregatedData = cleanAccounts(aggregatedData)
	if !isPlatformSearch(search) {
		aggregatedData = sortAccounts(aggregatedData, "website")
	}
	aggregatedData = truncateAccounts(aggregatedData)
	aggregatedData = sortAccounts(aggregatedData, "mrr")

	msg := common.FormatAccountInfos(aggregatedData, search)
	return json.Marshal(msg)
}

// helper functions

func exists(id string, website string, data []*models.AccountInfo) (result bool, index int) {
	result = false
	index = -1
	for i, account := range data {
		if account.Website == website || account.SiteId == id && id != "unknown" {
			index = i
			result = true
			break
		}
	}
	return result, index
}

// cleaning account arrays

func truncateAccounts(accounts []*models.AccountInfo) []*models.AccountInfo {
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
