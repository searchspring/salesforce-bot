package metabase

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/grokify/go-metabase/metabase"
	"github.com/grokify/go-metabase/metabaseutil"
	mo "github.com/grokify/oauth2more/metabase"
	"github.com/searchspring/nebo/validator"
)

type DAO interface {
	QueryAll(query string) (metabase.DatasetQueryResultsData, error)
}

type DAOImpl struct {
	Client *metabase.APIClient
}

func NewDAO(mbURL string, mbUser string, mbPassword string, mbToken string) DAO {
	if validator.ContainsEmptyString(mbURL, mbUser, mbPassword, mbToken) {
		return nil
	}

	config := mo.Config{
		BaseURL:       mbURL,
		Username:      mbUser,
		Password:      mbPassword,
		SessionID:     mbToken,
		TLSSkipVerify: true,
	}

	apiClient, _, err := metabaseutil.NewApiClient(config)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	return &DAOImpl{
		Client: apiClient,
	}
}

func (s *DAOImpl) QueryAll(search string) (metabase.DatasetQueryResultsData, error) {
	sqlInfo := metabaseutil.SQLInfo{}
	err := json.Unmarshal([]byte("MYSQL REQUEST"), &sqlInfo)
	if err != nil {
		log.Println(err.Error())
		return metabase.DatasetQueryResultsData{}, nil
	}

	info, resp, err := metabaseutil.QuerySQL(s.Client, sqlInfo.DatabaseID, sqlInfo.SQL)
	if err != nil {
		log.Fatal(err)
		return metabase.DatasetQueryResultsData{}, nil
	} else if resp.StatusCode >= 300 {
		log.Println(fmt.Sprintf("STATUS_CODE [%v]", resp.StatusCode))
		return metabase.DatasetQueryResultsData{}, nil
	}

	data := info.Data

	return data, nil
}
