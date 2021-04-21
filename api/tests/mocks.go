package mocks

import (
	"github.com/nlopes/slack"
	"github.com/searchspring/nebo/salesforce"
	"github.com/simpleforce/simpleforce"
)

type SalesforceDAO struct {
	searchKey string
}

func (s *SalesforceDAO) NPSQuery(search string) ([]*salesforce.AccountInfo, error) {
	accounts := []*salesforce.AccountInfo{}
	account := &salesforce.AccountInfo{Manager: search, Active: "active", MRR: 0, FamilyMRR: 0}
	accounts = append(accounts, account)
	s.searchKey = search
	return accounts, nil
}

func (s *SalesforceDAO) GetSearchKey() string {
	return s.searchKey
}

func (s *SalesforceDAO) StructFromResult(search string, result *simpleforce.QueryResult) ([]*salesforce.AccountInfo, error) {
	return []*salesforce.AccountInfo{}, nil
}
func (s *SalesforceDAO) Query(search string) ([]byte, error)   { return []byte{}, nil }
func (s *SalesforceDAO) IDQuery(search string) ([]byte, error) { return []byte{}, nil }
func (s *SalesforceDAO) ResultToMessage(search string, result *simpleforce.QueryResult) ([]byte, error) {
	return []byte{}, nil
}

type SlackDAO struct {
	Recorded []string
}

func (s *SlackDAO) SendSlackMessage(token string, attachments slack.Attachment, channel string) error {
	s.Recorded = []string{token, channel}
	return nil
}

func (s *SlackDAO) GetValues() []string {
	return s.Recorded
}