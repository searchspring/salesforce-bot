package aggregate

import (
	"github.com/searchspring/nebo/common"
	"github.com/searchspring/nebo/dals/metabase"
	"github.com/searchspring/nebo/dals/salesforce"
)


type Deps struct {
	MetabaseDAO metabase.DAO
	SalesforceDAO salesforce.DAO
}

type AggregateService interface {
	Query(query string) ([]byte, error)
}

type AggregateServiceImpl struct {
	Deps *Deps
}

func (d *AggregateServiceImpl) Query(search string) ([]byte, error) {
	metabaseData, err := d.deps.MetabaseDAO.Query(search)
	if err != nil {

	}
	salesforceData, err := d.deps.SalesforceDAO.Query(search)
	if err != nil {
		
	}

	var aggregatedData common.AccountInfo
	
	for i := 0; i < larestArrayLength(metabaseData, salesforceData); i++ {
		
	}

}


// helper functions 

func larestArrayLength(arr1 []*common.AccountInfo, arr2 []*common.AccountInfo) int {
	if len(arr1) > len(arr2) {
		return len(arr1)
	} else {
		return len(arr2)
	}
}
