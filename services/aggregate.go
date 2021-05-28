package aggregate

import (
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
	deps *Deps
}

func (d *AggregateServiceImpl) Query(query string) ([]byte, error) {
	
}


// helper functions 
