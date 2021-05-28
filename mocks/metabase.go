package mocks

import (
	"github.com/searchspring/nebo/dals/metabase"
	"github.com/searchspring/nebo/common"
	mb "github.com/grokify/go-metabase/metabase"
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
	return &metabase.NpsInfo{}, nil
}

func (s *MetabaseDAO) Query(search string) ([]*common.AccountInfo, error) { 
	response := []*common.AccountInfo{}
	return response, nil
}

func (s *MetabaseDAO) StructFromResult(result *mb.DatasetQueryResultsData) (*metabase.NpsInfo, error) {
	return &metabase.NpsInfo{}, nil
}

func (s *MetabaseDAO) ResultToMessage(search string, result *mb.DatasetQueryResultsData) ([]*common.AccountInfo, error) {
	response := []*common.AccountInfo{}
	return response, nil
}