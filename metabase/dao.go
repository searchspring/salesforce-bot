package metabase

import (
	"fmt"
	"log"

	"github.com/grokify/go-metabase/metabase"
	"github.com/grokify/go-metabase/metabaseutil"
	mo "github.com/grokify/oauth2more/metabase"
	//"github.com/searchspring/nebo/validator"
)

type DAO interface {
	QueryAll() (metabase.DatasetQueryResults, error)
}

type DAOImpl struct {
	Client *metabase.APIClient
}

const domainFields = "Name, Tracking_Code__c"

func NewDAO(mbURL string, mbUser string, mbPassword string, mbToken string) (DAO, mo.AuthResponse, error) {
	/*
	if validator.ContainsEmptyString(mbURL, mbUser, mbPassword, mbToken) {
		return nil, nil
	}
*/
	config := mo.Config{
		BaseURL:       mbURL,
		Username:      mbUser,
		Password:      mbPassword,
		SessionID:     mbToken,
		TLSSkipVerify: true,
	}

	apiClient, authInfo, err := metabaseutil.NewApiClient(config)
	if err != nil {
		log.Println(err.Error())
		return nil, mo.AuthResponse{}, err
	}

	return &DAOImpl{
		Client: apiClient,
	}, *authInfo, nil
}

func (s *DAOImpl) QueryAll() (metabase.DatasetQueryResults, error) {
	var databaseId int64 = 5
	q := "SELECT " + domainFields + " " + "FROM Websites"

	info, resp, err := metabaseutil.QuerySQL(s.Client, databaseId, q)
	if err != nil {
		log.Fatal(err)
		return metabase.DatasetQueryResults{}, nil
	} else if resp.StatusCode >= 300 {
		log.Println(fmt.Sprintf("STATUS_CODE [%v]", resp.StatusCode))
		return metabase.DatasetQueryResults{}, nil
	}

	return info, nil
}
