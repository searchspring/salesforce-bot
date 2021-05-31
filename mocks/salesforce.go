package mocks

import (
	"github.com/searchspring/nebo/common"
	"github.com/simpleforce/simpleforce"
)

type SalesforceDAO struct {
	searchKey string
}

func (s *SalesforceDAO) GetSearchKey() string { return s.searchKey }
func (s *SalesforceDAO) Query(search string) ([]*common.AccountInfo, error)   { return []*common.AccountInfo{}, nil }
func (s *SalesforceDAO) IDQuery(search string) ([]*common.AccountInfo, error) { return []*common.AccountInfo{}, nil }
func (s *SalesforceDAO) ResultToMessage(search string, result *simpleforce.QueryResult) ([]*common.AccountInfo, error) {return []*common.AccountInfo{}, nil}
