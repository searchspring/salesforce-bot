package aggregate

import (
	"encoding/json"
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
	largestArray, smallestArray := orderArrays(metabaseData, salesforceData)
	var overlap bool
	for i := 0; i < len(largestArray); i++ {
		overlap = false
		for k := 0; k < len(smallestArray); k++ {
			if largestArray[i].Website == smallestArray[k].Website || largestArray[i].SiteId == smallestArray[k].SiteId {
				overlap = true
				aggregatedData = append(aggregatedData, smallestArray[k])
				break
			} else {
				for m := 0; m < len(largestArray); m++ {
					if smallestArray[k] == largestArray[m] {
						aggregatedData = append(aggregatedData, smallestArray[k])
					}
				}
			}
		}
		if !overlap {
			aggregatedData = append(aggregatedData, largestArray[i])
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

func orderArrays(arr1 []*common.AccountInfo, arr2 []*common.AccountInfo) ([]*common.AccountInfo, []*common.AccountInfo) {
	if len(arr1) > len(arr2) {
		return arr1, arr2
	} else {
		return arr2, arr1
	}
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
