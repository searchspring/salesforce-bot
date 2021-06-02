package mocks

import (
	mb "github.com/grokify/go-metabase/metabase"
	"github.com/searchspring/nebo/dals/metabase"
	"github.com/searchspring/nebo/models"
)

type MetabaseDAO struct {
	searchKey string
}

func (s *MetabaseDAO) QueryAll() ([]byte, error) {
	response := []byte{}
	return response, nil
}

func (s *MetabaseDAO) GetSearchKey() string {
	return s.searchKey
}

func (s *MetabaseDAO) QueryNPS(search string) (*metabase.NpsInfo, error) {
	s.searchKey = search
	return &metabase.NpsInfo{
		Manager:   "tester",
		MRR:       1,
		FamilyMRR: 1,
	}, nil
}

func (s *MetabaseDAO) Query(search string) ([]*models.AccountInfo, error) {
	response := []*models.AccountInfo{}
	return response, nil
}

func (s *MetabaseDAO) StructFromResult(result *mb.DatasetQueryResultsData) (*metabase.NpsInfo, error) {
	return &metabase.NpsInfo{}, nil
}

func (s *MetabaseDAO) ResultToMessage(search string, result *mb.DatasetQueryResultsData) ([]*models.AccountInfo, error) {
	response := []*models.AccountInfo{}
	return response, nil
}
