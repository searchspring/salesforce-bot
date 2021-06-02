package mocks

import (
	"github.com/searchspring/nebo/models"
	"github.com/simpleforce/simpleforce"
)

type SalesforceDAO struct {
	searchKey string
}

func (s *SalesforceDAO) GetSearchKey() string { return s.searchKey }
func (s *SalesforceDAO) Query(search string) ([]*models.AccountInfo, error) {
	return []*models.AccountInfo{}, nil
}
func (s *SalesforceDAO) ResultToMessage(search string, result *simpleforce.QueryResult) ([]*models.AccountInfo, error) {
	return []*models.AccountInfo{}, nil
}
