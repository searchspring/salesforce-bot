package aggregate

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/searchspring/nebo/common"
	"github.com/searchspring/nebo/dals/metabase"
	"github.com/searchspring/nebo/dals/salesforce"
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

	var aggregatedData []*common.AccountInfo

	aggregatedData = append(aggregatedData, metabaseData...)

	for _, v := range salesforceData {
		for _, x := range aggregatedData {
			fmt.Printf("WebsiteSF: %s WebsiteMB: %s", v.Website, x.Website)
			fmt.Println()
			if !exists(v.SiteId, v.Website, aggregatedData) {
				fmt.Println("New Site Added: ", v.Website)
				aggregatedData = append(aggregatedData, v)
				break
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

func exists(id string, website string, data []*common.AccountInfo) (result bool) {
	result = false
	for _, account := range data {
		if account.SiteId == id || account.Website == website {
			result = true
			break
		}
	}
	return result
}

// cleaning account arrays

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

func sortAccounts(accounts []*common.AccountInfo, sortType string) []*common.AccountInfo {
	sort.Slice(accounts, func(i, j int) bool {
		if sortType == "website" {
			return len(accounts[i].Website) < len(accounts[j].Website)
		} else {
			return accounts[i].MRR > accounts[j].MRR
		}
	})
	return accounts
}
